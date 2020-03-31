# Alarm monitoring microservice
 
## Setting up and Launch
To setup local environment MongoDB and RabbitMQ instances are required. 
You may use docker-compose.yml from repo

1. Clone the repository.
2. Install dependencies, create the config file.
3. Run Database and Message Broker
4. Launch the project.

```bash
git clone https://github.com/eugeneverywhere/canopsis-test.git && cd canopsis-test
make setup && make configure 
docker-compose up -d
make run
```
## Tests
To run unit tests:
```bash
make test
```

To run utility for testing, which will send messages to rabbitMQ:
```bash
make ft:run
```
