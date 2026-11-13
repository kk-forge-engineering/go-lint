module example.com/app

go 1.26.0

require (
	github.com/gin-gonic/gin v1.0.0
	github.com/go-chi/chi/v5 v5.0.0
	github.com/gofiber/fiber/v2 v2.0.0
	github.com/gorilla/mux v1.0.0
	github.com/julienschmidt/httprouter v1.0.0
	github.com/labstack/echo/v4 v4.0.0
)

// The modules under third_party stand in for the real frameworks, so the fixture loads without the network.
replace (
	github.com/gin-gonic/gin => ./third_party/gin
	github.com/go-chi/chi/v5 => ./third_party/chi
	github.com/gofiber/fiber/v2 => ./third_party/fiber
	github.com/gorilla/mux => ./third_party/mux
	github.com/julienschmidt/httprouter => ./third_party/httprouter
	github.com/labstack/echo/v4 => ./third_party/echo
)
