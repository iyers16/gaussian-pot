FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Type-check before building — fail fast on TS errors.
RUN ./node_modules/.bin/tsc --noEmit

RUN npm run build

# ─── Runtime image ────────────────────────────────────────────────────────────
FROM node:20-alpine

ENV NODE_ENV=production
ENV HOSTNAME=0.0.0.0
ENV PORT=3000

WORKDIR /app

# Run as non-root for security.
RUN addgroup -g 1001 -S nodejs && adduser -S nextjs -u 1001
USER nextjs

COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static     ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/public           ./public

EXPOSE 3000

CMD ["node", "server.js"]
