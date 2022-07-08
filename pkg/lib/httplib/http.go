package httplib

import "bitbucket.org/noon-go/noonhttp"

var Client *noonhttp.ClientEntity

func InitializeHttp(client *noonhttp.ClientEntity) {
	Client = client
}
