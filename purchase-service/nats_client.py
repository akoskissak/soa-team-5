import asyncio
from nats.aio.client import Client as NATS
import json

nc = NATS()

async def connect_nats():
    if not nc.is_connected:
        await nc.connect("nats://localhost:4222")
    return nc

async def publish_event(subject: str, payload: dict):
    await connect_nats()
    await nc.publish(subject, json.dumps(payload).encode())

async def subscribe_event(subject: str, callback):
    await connect_nats()
    await nc.subscribe(subject, cb=callback)
