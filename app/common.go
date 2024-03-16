package app

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"standard-http-mongodb-storage/core/auth"
	"standard-http-mongodb-storage/core/responses"
	"strings"
	"time"
)

// checkPermission checks whether a permission is in the list of permissions.
func checkPermission(permission string, existingPermissions []interface{}) bool {
	for _, existingPermission := range existingPermissions {
		if perm, ok := existingPermission.(string); ok && (perm == "*" || perm == permission) {
			return true
		}
	}
	return false
}

// checkId ensures the :id is valid.
func checkId(ctx *gin.Context, arg string, raiseNotFoundOnError bool) (primitive.ObjectID, bool) {
	idParam := ctx.Param(arg)
	if id, err := primitive.ObjectIDFromHex(idParam); err != nil {
		if raiseNotFoundOnError {
			responses.NotFound(ctx)
		}
		return primitive.NilObjectID, false
	} else {
		return id, true
	}
}

// authenticate performs an authentication and permissions check.
func authenticate(ctx *gin.Context, collection *mongo.Collection, key, permission string) bool {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		responses.AuthMissing(ctx)
		return false
	}
	if !strings.HasPrefix(token, "Bearer ") {
		responses.AuthBadScheme(ctx)
		return false
	}

	token = token[7:]
	var tokenRecord auth.AuthToken
	if result := collection.FindOne(ctx, bson.M{
		"api-key": token, "valid_until": bson.M{
			"$not": bson.M{"$lt": time.Now()},
		},
	}); result.Err() != nil {
		responses.AuthNotFound(ctx)
		return false
	} else if err := result.Decode(&tokenRecord); err != nil {
		responses.InternalError(ctx)
		return false
	}

	hasPermission := false
	if globalPermissions, ok := tokenRecord.Permissions["*"]; ok {
		if globalPermissionsArray, ok := globalPermissions.([]interface{}); ok {
			hasPermission = checkPermission(permission, globalPermissionsArray)
		}
	}
	if !hasPermission {
		if localPermissions, ok := tokenRecord.Permissions[key]; ok {
			if localPermissionsArray, ok := localPermissions.([]interface{}); ok {
				hasPermission = checkPermission(permission, localPermissionsArray)
			}
		}
	}
	if !hasPermission {
		responses.AuthForbidden(ctx)
		return false
	}

	return true
}
