.PHONY: build-server build-worker build-notifier run-server run-worker run-notifier server worker notifier all
.SILENT: build-server build-worker build-notifier run-server run-worker run-notifier server worker notifier all

local:
	docker-compose up -d
 
docker-remove-server:
	cont=$(shell docker ps -a  | grep alert-server | awk '{print $$1 }'); \
	if [ "$$cont" ]; then docker stop alert-server && docker rm alert-server; \
	else echo "no container named alert-server found" ; \
	fi

docker-remove-worker:
	cont=$(shell docker ps -a  | grep alert-worker | awk '{print $$1 }'); \
	if [ "$$cont" ]; then docker stop alert-worker && docker rm alert-worker; \
	else echo "no container named alert-worker found" ; \
	fi

docker-remove-notifier:
	cont=$(shell docker ps -a  | grep alert-notifier | awk '{print $$1 }'); \
	if [ "$$cont" ]; then docker stop alert-notifier && docker rm alert-notifier; \
	else echo "no container named alert-notifier found" ; \
	fi;

build-server:
	docker build -f alert_initiator/Dockerfile -t alert-server .

build-worker:
	docker build -f alert_initiator/DockerfileWorker -t alert-worker .

build-notifier:
	docker build -f alert_notifier/Dockerfile -t alert-notifier .; \

run-server: docker-remove-server
	docker run -d --restart unless-stopped --cap-add=SYS_PTRACE -p 3335:3333 --network swillynetwork -v ${HOME}/.tmp:/tmp --name alert-server alert-server

run-worker: docker-remove-worker
	docker run -d --restart unless-stopped --cap-add=SYS_PTRACE --network swillynetwork -v ${HOME}/.tmp:/tmp --name alert-worker alert-worker

run-notifier: docker-remove-notifier
	docker run -d --restart unless-stopped --cap-add=SYS_PTRACE -p 3334:3333 --network swillynetwork -v ${HOME}/.tmp:/tmp --name alert-notifier alert-notifier

server: build-server run-server

worker: build-worker run-worker

notifier: build-notifier run-notifier

all: server worker notifier

