package routing

// internedStrings is a pool of commonly used parameter names to reduce
// string allocations during parameter extraction. Parameter names like
// "id", "userId", "name" are reused frequently across routes.
var internedStrings = map[string]string{
	"id":         "id",
	"userId":     "userId",
	"user_id":    "user_id",
	"name":       "name",
	"slug":       "slug",
	"postId":     "postId",
	"post_id":    "post_id",
	"commentId":  "commentId",
	"projectId":  "projectId",
	"orgId":      "orgId",
	"accountId":  "accountId",
	"itemId":     "itemId",
	"productId":  "productId",
	"categoryId": "categoryId",
}

// InternString returns an interned version of the string if it exists
// in the pool, otherwise returns the original string. This reduces
// allocations for commonly used parameter names.
func InternString(s string) string {
	if interned, ok := internedStrings[s]; ok {
		return interned
	}
	return s
}
