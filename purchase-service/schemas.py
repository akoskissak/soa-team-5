from pydantic import BaseModel
from typing import List, Optional

# --- OrderItem šeme ---
class OrderItemBase(BaseModel):
    tour_id: int
    tour_name: str
    price: float

class OrderItemCreate(OrderItemBase):
    pass

class OrderItem(OrderItemBase):
    id: int
    cart_id: int

    class Config:
        orm_mode = True # Omogućava Pydantic-u da čita podatke iz ORM objekata

# --- ShoppingCart šeme ---
class ShoppingCartBase(BaseModel):
    tourist_id: int

class ShoppingCartCreate(ShoppingCartBase):
    pass

class ShoppingCart(ShoppingCartBase):
    id: int
    total_price: float
    items: List[OrderItem] = []

    class Config:
        orm_mode = True