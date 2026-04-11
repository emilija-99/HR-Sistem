FROM node:20

WORKDIR /app

# copy only package files first (cache optimization)
COPY package*.json ./

RUN npm install

# now copy rest of project
COPY . .

EXPOSE 5173

CMD ["npm","run","dev","--","--host","0.0.0.0","--poll","2000"]
