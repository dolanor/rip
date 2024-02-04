SHELL = sh -xv

up:
	@go run . &

list:
	@curl -s localhost:8888/users/ | gojq .

new:
	@curl -s localhost:8888/users/ -X POST -H "Content-Type: application/json" -d '{ "name": "Philippe" }'

get:
	@curl -s localhost:8888/users/1 | gojq .

getxml:
	@curl -s localhost:8888/users/ -H "Accept: text/xml" | xmlfmt -f -


update:
	@curl -s localhost:8888/users/1 -X PUT -H "Content-Type: application/json" -d '{ "id": 1, "name": "Kamel" }' | gojq .

getfield:
	@curl -s localhost:8888/users/1/name | gojq .
	@curl -s localhost:8888/users/1/emailaddress | gojq .

updatefield:
	@curl -s localhost:8888/users/1/emailaddress -X PUT -H "Content-Type: application/json" -d '"kamel@example.com"' | gojq .

delete:
	@curl -s localhost:8888/users/1 -X DELETE
