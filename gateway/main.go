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
	"time"

	_ "github.com/lib/pq"
)

// Estructura para almacenar cada una de las rutas de nuestro JSON. `json:"prefix"` y `json:"target"`
// son metadatos de las variables Prefix y Target que les indican que atributo del JSON tomar
type Route struct {
	Prefix string `json:"prefix"`
	Target string `json:"target"`
}

// Estructura que representará nuestro gateway. Esta contiene un arreglo con las rutas extraidas del JSON,
// un mapa con los reverse proxies para cada servicio, así cómo un mapa con el estado de todos los backends
// con un mutex para regular el acceso concurrente. También contiene un semaforo para regular cuantas peticiones
// pueden procesarse a la vez
type Gateway struct {
	routes       []Route
	proxies      map[string]*httputil.ReverseProxy
	healthMutex  sync.RWMutex
	healthStatus map[string]bool
	sem          chan struct{}
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
	// Creamos variables donde guardaremos los mapas de nuestros reverse proxies y el estado de los backends
	proxies := make(map[string]*httputil.ReverseProxy)
	health := make(map[string]bool)

	for _, route := range routes {
		// Leemos las rutas previamente cargadas y las convertimos en una variable de tipo URL que Go entienda
		target, err := url.Parse(route.Target)
		if err != nil {
			log.Fatalf("invalid target url %s: %v", route.Target, err)
		}
		// Creamos un reverse proxy para cada backend
		proxies[route.Prefix] = httputil.NewSingleHostReverseProxy(target)
		health[route.Target] = false
	}

	// Devuelves el gateway con las rutas y reverse proxies ya cargados. Así pues, regresamos un lugar para
	// ir registrando los estados de los backends y semaforo que controlará el número máximo de peticiones
	return &Gateway{
		routes:       routes,
		proxies:      proxies,
		healthStatus: health,
		sem:          make(chan struct{}, maxPetitions),
	}
}

// ServeHTTP == función encargada de manejar las peticiones hechas al gateway
// Antes, handleRequest se encargaba de fabricar un handler para cada backend
func (gateway *Gateway) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// Cuando una nueva petición llega, le damos un espacio en el semaforo.
	// Cuando termine la petición, sin importar cómo, nos aseguramos de liberar el espacio
	gateway.sem <- struct{}{}
	defer func() { <-gateway.sem }()

	// Buscamos en nuestros proxies el que corresponda a la ruta de la petición.
	for prefix, proxy := range gateway.proxies {
		// Si alguno de los proxies corresponde a la ruta, el gateway le redirige la petición
		if strings.HasPrefix(request.URL.Path, prefix) {
			log.Printf("[%s] %s -> matched prefix %q", request.Method, request.URL.Path, prefix)
			proxy.ServeHTTP(writer, request)
			return
		}
	}
	// Si no, el gateway responde error 404
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
			for _, routes := range gateway.routes {
				waitgroup.Add(1)
				// Por cada ruta, checamos su estado (healthy/not healthy) y lo guardamos en la gateway
				go func(target string) {
					defer waitgroup.Done()
					isOk := checkHeartbeat(target, interval)
					gateway.healthMutex.Lock()
					gateway.healthStatus[target] = isOk
					gateway.healthMutex.Unlock()
				}(routes.Target)
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
	gateway.healthMutex.RLock()
	defer gateway.healthMutex.RUnlock()
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(gateway.healthStatus)
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
