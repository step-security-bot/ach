package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/moov-io/ach/server"
)

/**
CreateFile
curl -d '{"id":"08B751B2","immediateDestination":"9876543210", "immediateDestinationName":"Federal Reserve Bank", "immediateOrigin":"1234567890","immediateOriginName":"My Bank Name"}' -H "Content-Type: application/json" -X POST http://localhost:8080/files/

GetFile
curl -H "Content-Type: application/json" -X GET http://localhost:8080/files/1234

GetFiles
curl -H "Content-Type: application/json" -X GET http://localhost:8080/files/

DeleteFile
curl -H "Content-Type: application/json" -X DELETE http://localhost:8080/files/1234

CreateBatch
curl -d '{"id":"54321","serviceClassCode":220,"standardEntryClassCode":"WEB","companyName":"Your Company inc","companyIdentification":"121042882","companyEntryDescription":"Online Order","ODFIIdentification":"12104288"}' -H "Content-Type: application/json" -X POST http://localhost:8080/files/08B751B2/batches/

GetBatch
curl -H "Content-Type: application/json" -X GET http://localhost:8080/files/08B751B2/batches/54321

GetBatches
curl -H "Content-Type: application/json" -X GET http://localhost:8080/files/08B751B2/batches/

DeleteBatch
curl -H "Content-Type: application/json" -X DELETE http://localhost:8080/files/08B751B2/batches/54321
**/

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var s server.Service
	{
		s = server.NewService(server.NewRepositoryInMemory())
		s = server.LoggingMiddleware(logger)(s)
	}

	var h http.Handler
	{
		h = server.MakeHTTPHandler(s, log.With(logger, "component", "HTTP"))
	}

	// Listen for application termination.
	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)
	}()

	logger.Log("exit", <-errs)
}
