package functions

import (
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func init() {
	tp, err := InitTracing()
	if err != nil {
		group := slog.Group("init", slog.Group("InitTracing"))
		slog.Error("Failed to initialize tracing: %v", err, group)

		// If tracing fails to initialize, the program should exit.
		panic(err)
	}
	handler := InstrumentedHandler("chat", chatWatcher, tp)
	functions.HTTP("chat", handler)
}

func chatWatcher(w http.ResponseWriter, r *http.Request) {
	// Cache environment variables
	// Because the function is supposed to run on CloudFunctions, it is necessary to read the environment variables here.
	ytApiKey := os.Getenv("YOUTUBE_API_KEY")
	if ytApiKey == "" {
		slog.Error("YOUTUBE_API_KEY is not set")
		w.WriteHeader(http.StatusInternalServerError)
		panic("YOUTUBE_API_KEY is not set")
	}

	// Initialize span
	span := getSpanQuery(r.URL)

	// Create YouTube service
	ytSvc, err := youtube.NewService(r.Context(), option.WithAPIKey(ytApiKey))
	if err != nil {
		slog.Error("Failed to create YouTube service: %v", err)
		return
	}
	slog.Info("chatWatcher")
}

func getSpanQuery(u *url.URL) int {
	group := slog.Group("getSpanQuery")
	// Default value is 60 minutes
	// Because the update timing of the service group in the upper layer is every 60 minutes
	defVal := 60

	// Check if the request has a query parameter named "span"
	// If it does, return the value of the parameter as an integer
	// If it does not, return default value
	// If the value is not a number, return default value

	// Get the value of the query parameter named "span"
	span := u.Query().Get("span")
	if span == "" {
		slog.Info("span is empty", group)
		return defVal
	}

	// Convert the value to an integer
	spanInt, err := strconv.Atoi(span)
	if err != nil {
		slog.Error("Failed to set span because of invalid value", group)
		return defVal
	}

	// Return the value
	return spanInt
}
