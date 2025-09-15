import uvicorn
from fastapi import FastAPI

# Kreiranje instance aplikacije
app = FastAPI()

# Definišemo "health check" rutu da proverimo da li servis radi
@app.get("/")
def read_root():
    return {"status": "Purchase service is running"}

# Primer rute za dodavanje ture u korpu
# Kasnije ćemo ovde dodati pravu logiku
@app.post("/api/shopping-cart")
def add_to_cart():
    # TODO: Implementirati logiku dodavanja u korpu
    return {"message": "Item added to cart successfully"}

# Primer rute za checkout
@app.post("/api/checkout")
def checkout():
    # TODO: Implementirati logiku za checkout
    return {"message": "Checkout successful, tokens generated"}

# Pokretanje servera
if __name__ == "__main__":
    # 0.0.0.0 je bitno da bi bilo dostupno unutar Docker mreže
    uvicorn.run(app, host="0.0.0.0", port=8088)