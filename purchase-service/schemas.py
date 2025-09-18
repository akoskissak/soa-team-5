from pydantic import BaseModel
from typing import List, Optional
import uuid 
from datetime import datetime 

class OrderItemBase(BaseModel):
    tour_id: uuid.UUID
    tour_name: str
    price: float

class OrderItemCreate(OrderItemBase):
    pass

class OrderItem(OrderItemBase):
    id: int
    cart_id: int

    class Config:
        from_attributes = True

class ShoppingCartBase(BaseModel):
    tourist_id: str 

class ShoppingCartCreate(ShoppingCartBase):
    pass

class ShoppingCart(ShoppingCartBase):
    id: int
    total_price: float
    items: List[OrderItem] = []

    class Config:
        from_attributes = True

class TourPurchaseTokenBase(BaseModel):
    tour_id: uuid.UUID
    tourist_id: str 
    token: uuid.UUID
    tour_name: str
    price: float

class TourPurchaseToken(TourPurchaseTokenBase):
    id: int
    
    class Config:
        from_attributes = True