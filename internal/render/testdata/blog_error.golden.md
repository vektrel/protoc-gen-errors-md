# api.blog.v1 error codes

> Default HTTP code: `500` (any reason not listed below uses this code).

| Reason | HTTP | Description |
|---|---|---|
| `UNAUTHORIZED` | 401 |  |
| `FORBIDDEN` | 403 |  |
| `ARTICLE_NOT_FOUND` | 404 | Article not found or already deleted. |
| `TITLE_INVALID` | 400 | Title contains illegal characters such as \| or newline. |
| `RATE_LIMITED` | 429 | Comment rate limit exceeded. Please try again later. |
| `INTERNAL` | 500 |  |
