# Indique à Docker quelle version de syntaxe utiliser pour lire ce fichier
# syntax=docker/dockerfile:1

# --- ÉTAPE 1 : LA CONSTRUCTION (L'Atelier) ---
# On définit une variable pour la version de Go
ARG GO_VERSION=1.24.2
# On part d'une image contenant déjà tous les outils Go. On l'appelle "build"
FROM golang:${GO_VERSION} AS build

# On définit le dossier de travail à l'intérieur de l'image
WORKDIR /src

# On installe les dépendances système nécessaires pour compiler du C et SQLite
RUN apt-get update && apt-get install -y \
    gcc \
    musl-dev \
    musl-tools \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/* # Nettoyage pour gagner de la place

# On télécharge les bibliothèques Go (dépendances)
# Les --mount servent à mettre en cache les téléchargements pour aller plus vite la prochaine fois
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# On copie tout ton code source du dossier actuel vers l'image
COPY . .

# On compile le programme Go en un fichier binaire nommé "/bin/server"
# Les options bizarres (CGO_ENABLED=1, static) servent à inclure SQLite directement 
# dans le fichier pour qu'il soit autonome.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    CC=musl-gcc \
    CGO_ENABLED=1 \
    GOOS=linux \
    go build \
    -ldflags="-linkmode external -extldflags '-static'" \
    -tags sqlite_omit_load_extension \
    -o /bin/server .

################################################################################
# --- ÉTAPE 2 : L'EXÉCUTION (Le Carton Final) ---
# On repart de zéro avec "Alpine", une image Linux ultra-légère (env. 5 Mo)
FROM alpine:latest AS final

# On installe les certificats de sécurité (HTTPS) et les fuseaux horaires
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

# Par sécurité, on crée un utilisateur "appuser" pour ne pas lancer l'appli en tant que "root" (admin)
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser

# ON RÉCUPÈRE UNIQUEMENT CE DONT ON A BESOIN de l'étape "build"
COPY --from=build /bin/server /bin/
COPY --from=build /src/templates /templates/
COPY --from=build /src/static /static/
COPY --from=build /src/data/schemaRideUp.sql /data/schemaRideUp.sql

# On crée un dossier pour la base de données et on donne les droits à notre utilisateur
RUN mkdir -p /data && chown -R appuser:appuser /data

# On dit à Docker d'utiliser notre utilisateur limité
USER appuser

# On indique que l'application écoute sur le port 5090
EXPOSE 5090

# La commande qui se lance au démarrage du conteneur
ENTRYPOINT [ "/bin/server" ]