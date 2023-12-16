# mmk-be

## MMK Back-End

Money Manager Keeper is a web application that allows users to track their expenses and incomes. It also provides a
visual representation of the user's financial status.

### How to run the application:

1. Clone the repository
2. Install the dependencies with `go get all`
3. Create a `.env` file based on the `.env.example` file
4. Run the application with `go run main.go`

The application will be available at `localhost:2000`

### Architecture

The application is built using the MVC architecture. The `main.go` file is the entry point of the application. To add
more service, controller or model files, create a new folder in the respective directory and add the files. Protobuf
files are used to define the models and services. To generate protobuf files,
run `protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative <file>.proto`
in the respective directory. The generated proto should be in `pb` directory.



