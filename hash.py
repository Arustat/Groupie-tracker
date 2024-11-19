from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.backends import default_backend
import os

# Exemple de clé et de données chiffrées
key = bytes.fromhex('a1c299e572ff8c643a857d3fdb3e5c7c10101010101010101010101010101010')  # Remplace avec ta clé
iv = os.urandom(16)  # Utilise le vecteur d'initialisation approprié
ciphertext = bytes.fromhex('aad3b435b51404eeaad3b435b51404ee:2b87e7c93a3e8a0ea4a581937016f341')  # Remplace avec ton texte chiffré

# Déchiffrement
cipher = Cipher(algorithms.AES(key), modes.CBC(iv), backend=default_backend())
decryptor = cipher.decryptor()
plaintext = decryptor.update(ciphertext) + decryptor.finalize()

print(plaintext)
