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
opentelemetry-instrument \
    --traces_exporter console,otlp \
    --metrics_exporter none \
    --service_name test \
    --exporter_otlp_endpoint 0.0.0.0:4317 \
    python myapp.py

```
