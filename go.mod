module github.com/Templer90/Scrumpoker

go 1.13

require github.com/gorilla/mux v1.8.0

require internal/handler v1.0.0

replace internal/handler => ./internal/pkg/handler

require internal/models v1.0.0

replace internal/models => ./internal/pkg/models

require internal/util v1.0.0

replace internal/util => ./internal/pkg/util
