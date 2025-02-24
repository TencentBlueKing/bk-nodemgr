.PHONY: default all clean

default: all

application/application.pb.go: application/application.proto
	@protoc --go_opt=paths=source_relative --go_out=. application/application.proto

backend/backend.pb.go: backend/backend.proto
	@protoc --go_opt=paths=source_relative --go_out=. backend/backend.proto

all: application/application.pb.go backend/backend.pb.go
	@echo "successfully generated protocols"

clean:
	@rm -f application/application.pb.go
	@rm -f backend/backend.pb.go
