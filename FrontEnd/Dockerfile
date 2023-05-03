# Utilise l'image officielle de Node.js en version 18
FROM node:18

# Expose le port 3000
EXPOSE 3000

# Définit le répertoire de travail
WORKDIR /app

# Copy local au container
COPY . .

# Installe les dépendances
RUN yarn

# Build l'application pour la production
RUN yarn build

# Démarre l'application
CMD [ "yarn", "preview" ]