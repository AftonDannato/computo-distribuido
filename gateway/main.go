package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
)

// Estructura para almacenar cada una de las rutas de nuestro JSON. `json:"prefix"` y `json:"target"`
// son metadatos de las variables Prefix y Target que les indican que atributo del JSON tomar
type Route struct {
	Prefix  string   `json:"prefix"`
	Targets []string `json:"targets"`
}

// Estructura que representará nuestro gateway. Esta contiene un arreglo con las rutas extraidas del JSON,
// un mapa con los reverse proxies para cada servicio, así cómo un mapa con el estado de todos los backends
// con un mutex para regular el acceso concurrente. También contiene un semaforo para regular cuantas peticiones
// pueden procesarse a la vez
type Gateway struct {
	groups []*RouteGroup
	sem    chan struct{}
}

// Debido a que ahora hay más de un backend, creamos una estructura para agrupar su URL, su estado y
// su reverse proxy, en lugar de tenerlos almacenados en el gateway directamente
// Cada backend es el unico que accede a su estatus, por lo que no necesitamos protegerlo con un mutex
type Backend struct {
	URL          *url.URL
	Proxy        *httputil.ReverseProxy
	healthStatus atomic.Bool
}

// Estructura en la que agrupamos un prefijo con los backends que pueden atender a las peticiones que se le hagan
type RouteGroup struct {
	Prefix       string
	Backends     []*Backend
	counter      int
	counterMutex sync.Mutex
}

// Obtiene el siguiente índice del round robin y avanza el contador de manera concurrentemente segura.
func (routeGroup *RouteGroup) nextIndex(n int) int {
	routeGroup.counterMutex.Lock()
	defer routeGroup.counterMutex.Unlock()
	idx := routeGroup.counter
	routeGroup.counter = (routeGroup.counter + 1) % n
	return idx
}

// Función que hace el round robin que escoge un backend sano para responder una petición entrante
func (routeGroup *RouteGroup) chooseBackend() *Backend {
	n := len(routeGroup.Backends)
	// Buscamos en nuestro grupo de Backends el primero que esté sano, empezando por el backend siguiente
	// al backend que respondió la última petición al servicio.
	for i := 0; i < n; i++ {
		idx := routeGroup.nextIndex(n)
		b := routeGroup.Backends[idx]
		if b.healthStatus.Load() {
			return b
		}
	}
	// Si recorremos todo nuestro grupo y no se encontró ningún Backend sano, se indica devolviendo nil
	return nil
}

// Función para la carga de rutas en el JSON y conversión de estas a estructuras de tipo Route
func loadRoutes(path string) ([]Route, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var routes []Route
	if err := json.Unmarshal(data, &routes); err != nil {
		return nil, err
	}
	return routes, nil
}

// Función para crear el Gateway
func NewGateway(routes []Route, maxPetitions int) *Gateway {
	// Creamos un slice que contiene nuestros grupos de ruta
	var groups []*RouteGroup
	// Para cada ruta, iteramos sobre los posibles backends que pueden responder a ese prefix, y los agregamos
	// a un RouteGroupe, así el gateway tendrá un pool de backends en lugar de uno solo
	for _, route := range routes {
		routeGroup := &RouteGroup{Prefix: route.Prefix}
		for _, targetRoute := range route.Targets {
			// Leemos las rutas previamente cargadas y las convertimos en una variable de tipo URL que Go entienda
			targetURL, err := url.Parse(targetRoute)
			if err != nil {
				log.Fatalf("invalid target url %s: %v", targetRoute, err)
			}
			// Creamos una instancia de "Backend" que almacenará su dirección y el RP por el cual se comunica
			backend := &Backend{URL: targetURL, Proxy: httputil.NewSingleHostReverseProxy(targetURL)}
			routeGroup.Backends = append(routeGroup.Backends, backend)
		}
		// Cuando se obtienen todos los backends asociados a la ruta, se guarda el RouteGroup en el gatewy
		groups = append(groups, routeGroup)
	}

	// Devuelves el gateway con los grupos de backends. Así pues, regresamos un semaforo que controlará el
	// número máximo de peticiones
	return &Gateway{
		groups: groups,
		sem:    make(chan struct{}, maxPetitions),
	}
}

// ServeHTTP == función encargada de manejar las peticiones hechas al gateway
// Antes, handleRequest se encargaba de fabricar un handler para cada backend
func (gateway *Gateway) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// Cuando una nueva petición llega, le damos un espacio en el semaforo.
	// Cuando termine la petición, sin importar cómo, nos aseguramos de liberar el espacio
	gateway.sem <- struct{}{}
	defer func() { <-gateway.sem }()

	// Buscamos en nuestros routeGroups el que corresponda a la ruta de la petición.
	for _, routeGroup := range gateway.groups {
		// Mientras el prefix correspondiente al grupo actual no sea el de la petición, lo saltamos
		if !strings.HasPrefix(request.URL.Path, routeGroup.Prefix) {
			continue
		}
		chosenBackend := routeGroup.chooseBackend()
		// Si no encuentra ningún backend vivo, el gateway responde error 503
		if chosenBackend == nil {
			http.Error(writer, "No healthy backend available", http.StatusServiceUnavailable)
			return
		}
		log.Printf("[%s] %q -> matched prefix %q -> handled by: %s", request.Method, request.URL.Path,
			routeGroup.Prefix, chosenBackend.URL.Host)
		chosenBackend.Proxy.ServeHTTP(writer, request)
		return
	}
	// Si no encuentra ningun routeGroup que responda a esa petición, el gateway responde error 404
	http.NotFound(writer, request)
}

func (gateway *Gateway) startHeartbeatMonitor(interval time.Duration) {
	// Creamos un ticker, que mandará una señal por un channel cada n intervalo de tiempo
	ticker := time.NewTicker(interval)
	go func() {
		// Un for range sobre un channel lo que hará es quedarse esperando por las cosas que lleguen a
		// dicho canal, y cuando algo llegue, ejecutará el codigo contenido. Este canal se mantendrá
		// por toda la duración del programa.
		for range ticker.C {
			// Hacemos un grupo de espera para asegurarnos de que todos los health checks terminen antes
			// de empezar una nueva ronda
			var waitgroup sync.WaitGroup
			for _, groups := range gateway.groups {
				for _, backend := range groups.Backends {
					waitgroup.Add(1)
					// Por cada backend, checamos su estado (healthy/not healthy) y lo guardamos
					go func(backend *Backend) {
						defer waitgroup.Done()
						backend.healthStatus.Store(checkHeartbeat(backend.URL.String(), interval))
					}(backend)
				}
			}
			waitgroup.Wait()
		}
	}()
}

// Función que manda una petición a la ruta /heartbeat del backend. Si responde codigo 200, lo marcamos
// saludable, en caso contrario, lo marcamos no saludable
func checkHeartbeat(target string, interval time.Duration) bool {
	// Calculamos que el tiempo de timeout será la mitad del tiempo del ticker, con un máximo de 4
	// segundos de timeout.
	timeout := 4 * time.Second
	if interval < 8*time.Second {
		timeout = interval / 2
	}

	client := http.Client{Timeout: timeout}
	resp, err := client.Get(target + "/heartbeat")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Función del gateway que nos permite obtener y responder los estatus de los backends
func (gateway *Gateway) statusHandler(writer http.ResponseWriter, request *http.Request) {
	type backendStatus struct {
		Target  string `json:"target"`
		Healthy bool   `json:"healthy"`
	}
	// ya no quiero comentar, tengo sueño...
	out := make(map[string][]backendStatus)
	for _, routeGroup := range gateway.groups {
		var list []backendStatus
		for _, backend := range routeGroup.Backends {
			list = append(list, backendStatus{Target: backend.URL.String(), Healthy: backend.healthStatus.Load()})
		}
		out[routeGroup.Prefix] = list
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(out)
}

// Función que responde si el gateway está vivo
func heartbeatHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("gateway alive"))
}

// Hilo principal del programa, donde cargamos nuestras rutas, creamos nuestro gateway y lo ponemos
// a escuchar en localhost. Es el endpoint para el frontend
func main() {

	routes, err := loadRoutes("./routes.json")
	if err != nil {
		log.Fatal(err)
	}

	// eventModel := models.NewEventModel(db)

	gateway := NewGateway(routes, 10)

	gateway.startHeartbeatMonitor(5 * time.Second)

	router := http.NewServeMux()

	router.HandleFunc("/heartbeat", heartbeatHandler)
	router.HandleFunc("/status", gateway.statusHandler)
	router.Handle("/", gateway)

	log.Println("Servidor escuchando en http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
