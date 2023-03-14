from quart import Quart, request
from opentelemetry.instrumentation.asgi import OpenTelemetryMiddleware
from opentelemetry.sdk.trace.export import (
    BatchSpanProcessor,
    ConsoleSpanExporter
)
from opentelemetry.sdk.resources import Resource, ResourceAttributes
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk import trace
from grpc import Compression


provider = trace.TracerProvider(
    resource=Resource.create({
        ResourceAttributes.SERVICE_NAME: "quart",
        ResourceAttributes.DEPLOYMENT_ENVIRONMENT: "staging"
        
    })
)

# processor = BatchSpanProcessor(ConsoleSpanExporter())
processor = BatchSpanProcessor(
    OTLPSpanExporter("localhost:4317",
                     insecure=True,
                     compression=Compression.Gzip
                     ))

provider.add_span_processor(processor)

app = Quart('quart')

def default_span_details(scope: dict):
    routeName = scope.get("path")
    if routeName == "" :
        routeName = "unknown"
    span_name = routeName
    route = routeName

    return span_name, {
        "http.route": route,
    }

    

app.asgi_app = OpenTelemetryMiddleware(app.asgi_app,
                                       tracer_provider=provider,
                                       default_span_details=default_span_details)


@app.route("/")
async def echo():
    print(request.is_json, request.mimetype_params)
    data = await request.get_json()
    return {"input": data, "extra": True}


@app.route("/remote")
async def remote():
    print(request.is_json, request.mimetype_params)
    data = await request.get_json()
    return {"input": data, "extra": True}

if __name__ == "__main__":
    app.run(
        port=8080,
        debug=True,
    )
