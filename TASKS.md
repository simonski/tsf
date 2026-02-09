Implement the testing according to the testing description in rules, specifically

- a target for unit tests `make test-unit`
    This will run go unit tests only, does not make an calls to external components.

- a target for integration tests `make test-integration`
    This will setup a clean test database
    This will run every client API call via a go test suite

- a target for frontend tests `make test-frontend`
    This will setup a clean test database
    This will run a javascript test framework to 
        test every API call
        test all website functionality

- a single target for all tests `make test`
    This will run in sequence unit, integration and frontend tests.

initdb:
    Introduce a THIRD double-hyphen option, "--populate", which 
    will then populate the database with the contents of the scripts/initdb folders:
        create users as per scripts/initdb/users.md
        create projects as per scripts/initdb/projects.md
        create workers as per scripts/initdb/workers.md
        create orchestrators as per scripts/initdb/orchestrators.md    
        create roles as per scripts/initdb/roles.md
