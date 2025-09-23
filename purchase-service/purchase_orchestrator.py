import asyncio
import json
from typing import List
from fastapi import APIRouter
from nats.aio.client import Client as NATS
import models
from sqlalchemy import func
from sqlalchemy.orm import Session

router = APIRouter()

# Orkestrator SAGA
class PurchaseOrchestrator:
	def __init__(self, nc: NATS):
		self.nc = nc
		self.pending = {}

	async def subscribe(self):
		await self.nc.subscribe("purchase_reply", cb=self.handle_payment_reply)

	async def startCheckout(self, tourist_id: str, tokens: List[models.TourPurchaseToken], db: Session):
		total_amount = sum(token.price for token in tokens)
		event = {
			"userId": tourist_id,
			"amount": total_amount,
			"command": "SUBTRACT"
		}

		loop = asyncio.get_running_loop()
		future = loop.create_future()
		self.pending[tourist_id] = {"future": future, "db": db}

		await self.nc.publish("purchase_publish", json.dumps(event).encode())

		result = await future
		return result

	async def handle_payment_reply(self, msg):
		data = json.loads(msg.data.decode())
		user_id = data["userId"]
		status = data["status"]  # "COMPLETED" ili "FAILED"
		amount = data["amount"]

		entry = self.pending.pop(user_id, None)
		if not entry:
			print(f"Nema pending entry za user {user_id}, ignorisem reply.")
			return

		future = entry["future"]
		db: Session = entry["db"]

		if status == "FAILED":
			# AKO JE NEUSPESNO OBRISI TOKENE KORISNIKU IZ db
			print(f"Transaction failed. Deleting tokens for user {user_id}...")
			latest_time = db.query(func.max(models.TourPurchaseToken.created_at)).filter(
				models.TourPurchaseToken.tourist_id == user_id
			).scalar()

			tokens_to_delete = db.query(models.TourPurchaseToken).filter(
				models.TourPurchaseToken.tourist_id == user_id,
				models.TourPurchaseToken.created_at == latest_time
			).all()

			for token in tokens_to_delete:
				db.delete(token)
			db.commit()
			result = []

			print(f"Rollback: deleted tokens for user {user_id}")

		else:
			print(f"Primljena poruka da je skinuto {amount} sa racuna")
			result = db.query(models.TourPurchaseToken).filter(
            models.TourPurchaseToken.tourist_id == user_id
        ).all()
		
		if not future.done():
			future.set_result(result)
		print(f"CHECKOUT ZAVRSEN za user {user_id} sa statusom {status}.")