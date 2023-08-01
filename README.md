# opentel
Um simple projeto para a instrumentalizacao 



```shell
docker-compose up -d
```

The demo exposes the following backends:

- Jaeger at http://0.0.0.0:16686
- Zipkin at http://0.0.0.0:9411
- Prometheus at http://0.0.0.0:9090 

```shell

npm install --save @opentelemetry/auto-instrumentations-node                     02:10:27 
npm install --save @opentelemetry/exporter-trace-otlp-http
npm install --save @opentelemetry/resources
npm install --save @opentelemetry/sdk-node
npm install --save @opentelemetry/semantic-conventions

```
