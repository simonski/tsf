# Deploy 1 - Local non-docker

make server
    run the server and frontend on 8080, run the orchestrator included

make local-caddy    
    run local caddy on 443 that reverse proxies 
        / to frontend
        /api/* to local server backend

# Deploy 2 - Local docker

make docker
    assemble the server binary

make docker-up-local

    run all docker components via a docker-compose.  use a volume to hold the sqlite file between up-and-down calls
        container: server backend runnind on 8080 
        frontend: javascript, html, css served statically on 8080
        orchestrator: run the orchestrator in a different container that points to the server as the TASK_URL
        proxy: caddy listening on https that routes 
            / to frontend
            /api to backend 8080

make docker-down-local

    stops all local docker images

