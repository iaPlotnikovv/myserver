
## go application
FROM golang
## We create an /app directory within our
## image that will hold our application source
## files
RUN mkdir /serv
## We copy everything in the root directory
## into our /app directory
ADD . /serv
## We specify that we now wish to execute 
## any further commands inside our /app
## directory
WORKDIR /serv
## we run go build to compile the binary
## executable of our Go program
RUN go build  ./cmd/main.go
## Our start command which kicks off
## our newly created binary executable
CMD ["/serv/main"]

