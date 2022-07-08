FROM golang AS builder

#This env var needs to be set for Gomodule
ENV GO111MODULE=on
ENV GOPRIVATE=bitbucket.org/noon-micro


#Creating the working directory
WORKDIR /curriculum
#Run the go mod download if only the changes occured in either of these 2 files
RUN git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/"
COPY go.mod .
COPY go.sum .
RUN mkdir -p /root/.ssh
COPY /build/id_rsa /root/.ssh/
RUN ssh-keyscan bitbucket.org >> /root/.ssh/known_hosts
RUN go mod download
#Copy your code to this working directory
COPY . .
#This will build the go server binary files
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build /curriculum/cmd/curriculum

#This is to spin up the base image to run the binaries of go
FROM alpine
RUN echo $SERVICE
COPY --from=builder /curriculum/* /curriculum/
EXPOSE 8080

COPY /build/start.sh /build/
COPY /.dockerignore /
RUN chmod 777 /.dockerignore
RUN chmod 777 /build/start.sh

ENTRYPOINT ["sh","/build/start.sh"]

