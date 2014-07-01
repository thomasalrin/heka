
AMQPOutput
==========

Connects to a remote AMQP broker (RabbitMQ) and sends messages to the
specified queue. The message is serialized if specified, otherwise only
the raw payload of the message will be sent. As AMQP is dynamically
programmable, the broker topology needs to be specified.

Config:

- URL (string):
    An AMQP connection string formatted per the `RabbitMQ URI Spec
    <http://www.rabbitmq.com/uri-spec.html>`_.
- Exchange (string):
    AMQP exchange name
- Queue (string):
    AMQP queue name
- ExchangeType (string):
    AMQP exchange type (`fanout`, `direct`, `topic`, or `headers`).
- ExchangeDurability (bool):
    Whether the exchange should be configured as a durable exchange. Defaults
    to non-durable.
- ExchangeAutoDelete (bool):
    Whether the exchange is deleted when all queues have finished and there
    is no publishing. Defaults to auto-delete.
- RoutingKey (string):
    The message routing key used to bind the queue to the exchange. Defaults
    to empty string.
- Persistent (bool):
    Whether published messages should be marked as persistent or transient.
    Defaults to non-persistent.

.. versionadded:: 0.6

- ContentType (string):
     MIME content type of the payload used in the AMQP header. Defaults to
     "application/hekad".
- Encoder (string)
    Default to "ProtobufEncoder".

- Queue (string):
    Name of the queue to consume from, an empty string will have the broker
    generate a name for the queue. Defaults to empty string.
- QueueDurability (bool):
    Whether the queue is durable or not. Defaults to non-durable.
- QueueExclusive (bool):
    Whether the queue is exclusive (only one consumer allowed) or not.
    Defaults to non-exclusive.
- QueueAutoDelete (bool):
    Whether the queue is deleted when the last consumer un-subscribes.
    Defaults to auto-delete.
- QueueTTL (int):
    Allows ability to specify TTL in milliseconds on Queue declaration for
    expiring messages. Defaults to undefined/infinite.

.. versionadded:: 0.6

- tls (TlsConfig):
    An optional sub-section that specifies the settings to be used for any
    SSL/TLS encryption. This will only have any impact if `URL` uses the
    `AMQPS` URI scheme. See :ref:`tls`.

Example (that sends log lines from the logger):

.. code-block:: ini

    [AMQPOutput]
    url = "amqp://guest:guest@rabbitmq/"
    exchange = "testout"
    exchangeType = "fanout"
    message_matcher = 'Logger == "TestWebserver"'
