package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// IdempotencyMiddleware ensures duplicate requests with the same Idempotency-Key
// header receive identical responses without re-executing business logic.
func IdempotencyMiddleware(rdb *redis.Client, lockTTL, cacheTTL time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only check idempotency for state-mutating requests (POST, PUT, PATCH)
		if c.Method() != fiber.MethodPost && c.Method() != fiber.MethodPut && c.Method() != fiber.MethodPatch {
			return c.Next()
		}

		key := c.Get("Idempotency-Key")
		if key == "" {
			// If no idempotency key is provided, proceed normally
			return c.Next()
		}

		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}

		redisCacheKey := "idempotency:response:" + key
		redisLockKey := "idempotency:lock:" + key

		// 1. Check if cached response exists for this key
		cachedResp, err := rdb.Get(ctx, redisCacheKey).Bytes()
		if err == nil && len(cachedResp) > 0 {
			c.Set("X-Cache-Lookup", "HIT")
			c.Set("Content-Type", "application/json; charset=utf-8")
			return c.Status(http.StatusOK).Send(cachedResp)
		}

		// 2. Acquire atomic in-progress lock using SETNX
		acquired, err := rdb.SetNX(ctx, redisLockKey, "1", lockTTL).Result()
		if err != nil {
			// If Redis is temporarily unreachable, let request proceed to avoid hard blocking
			return c.Next()
		}

		if !acquired {
			// Another concurrent request is already being processed with this key
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"error": "A request with this Idempotency-Key is currently being processed",
			})
		}

		defer func() {
			// Release lock after completion
			_ = rdb.Del(context.Background(), redisLockKey)
		}()

		// 3. Process the actual request handler
		if err := c.Next(); err != nil {
			return err
		}

		// 4. Cache successful responses (HTTP 200 - 299)
		status := c.Response().StatusCode()
		if status >= 200 && status < 300 {
			respBody := c.Response().Body()
			_ = rdb.Set(context.Background(), redisCacheKey, respBody, cacheTTL).Err()
		}

		return nil
	}
}
