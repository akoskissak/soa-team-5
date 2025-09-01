package handlers

import (
	"errors"
	"follower-service/db"
	"follower-service/models"
	"follower-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const relF = "FOLLOWS"

func Follow(c *gin.Context) {
	claims, jwtfErr := utils.VerifyJWT(c)
	if jwtfErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": jwtfErr.Error()})
		return
	}

	if claims["role"] == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only guides and tourists can access this route"})
		return
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token claims"})
		return
	}

	var req models.FollowReq
	if err := c.ShouldBindJSON(&req); err != nil || req.To == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if username == req.To {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot follow yourself"})
		return
	}

	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(c)

	_, err := session.ExecuteWrite(c, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `MERGE (a:User {username:$from})
              MERGE (b:User {username:$to})
              MERGE (a)-[:` + relF + `]->(b)`
		_, err := tx.Run(c, q, map[string]any{"from": username, "to": req.To})
		return nil, err
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "followed"})
}

/*func Unfollow(c *gin.Context) {
	claims, jwtfErr := utils.VerifyJWT(c)
	if jwtfErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": jwtfErr.Error()})
		return
	}

	if claims["role"] == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only guides and tourists can access this route"})
		return
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token claims"})
		return
	}

	var req models.FollowReq
	if err := c.ShouldBindJSON(&req); err != nil || req.To == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if username == req.To {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot unfollow yourself"})
		return
	}

	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(c)

	_, err := session.ExecuteWrite(c, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `
		MATCH (a:User {username:$from})-[r:` + relF + `]->(b:User {username:$to})
		DELETE r
		RETURN COUNT(*)`
		res, err := tx.Run(c, q, map[string]any{"from": username, "to": req.To})
		if err != nil { return nil, err }
		if res.Next(c) {
			return res.Record().Values[0], nil
		}
		return nil, errors.New("no result")
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unfollowed", "from": username, "to": req.To})
}*/

func Unfollow(c *gin.Context) {
	claims, jwtfErr := utils.VerifyJWT(c)
	if jwtfErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": jwtfErr.Error()})
		return
	}

	if claims["role"] == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only guides and tourists can access this route"})
		return
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token claims"})
		return
	}

	toUsername := c.Param("to")
	if toUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if username == toUsername {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot unfollow yourself"})
		return
	}

	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(c)

	_, err := session.ExecuteWrite(c, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `
        MATCH (a:User {username:$from})-[r:` + relF + `]->(b:User {username:$to})
        DELETE r
        RETURN COUNT(*)`
		res, err := tx.Run(c, q, map[string]any{"from": username, "to": toUsername})
		if err != nil {
			return nil, err
		}
		if res.Next(c) {
			return res.Record().Values[0], nil
		}
		return nil, errors.New("no result")
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unfollowed", "from": username, "to": toUsername})
}

func GetFollowing(c *gin.Context) {
	u := c.Param("username")
	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(c)

	data, err := session.ExecuteRead(c, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `MATCH (:User {username:$u})-[:` + relF + `]->(f:User) RETURN f.username AS u ORDER BY u`
		res, err := tx.Run(c, q, map[string]any{"u": u})
		if err != nil {
			return nil, err
		}
		usernames := make([]string, 0)
		for res.Next(c) {
			usernames = append(usernames, res.Record().Values[0].(string))
		}
		return usernames, nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": u, "following": data})
}

func GetFollowers(c *gin.Context) {
	u := c.Param("username")
	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(c)

	data, err := session.ExecuteRead(c, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `MATCH (f:User)-[:` + relF + `]->(:User {username:$u}) RETURN f.username AS u ORDER BY u`
		res, err := tx.Run(c, q, map[string]any{"u": u})
		if err != nil {
			return nil, err
		}
		usernames := make([]string, 0)
		for res.Next(c) {
			usernames = append(usernames, res.Record().Values[0].(string))
		}
		return usernames, nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": u, "followers": data})
}

func Recommend(c *gin.Context) {
	claims, jwtfErr := utils.VerifyJWT(c)
	if jwtfErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": jwtfErr.Error()})
		return
	}

	if claims["role"] == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only guides and tourists can access this route"})
		return
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token claims"})
		return
	}
	limit := 10

	session := db.Driver.NewSession(c, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(c)

	data, err := session.ExecuteRead(c, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `
		MATCH (me:User {username:$u})-[:` + relF + `]->(m:User)-[:` + relF + `]->(rec:User)
		WHERE NOT (me)-[:` + relF + `]->(rec) AND me <> rec
		RETURN rec.username AS username, COUNT(DISTINCT m) AS mutuals
		ORDER BY mutuals DESC, username ASC
		LIMIT $limit`
		res, err := tx.Run(c, q, map[string]any{"u": username, "limit": limit})
		if err != nil {
			return nil, err
		}
		recs := make([]models.RecDTO, 0)
		for res.Next(c) {
			rec := models.RecDTO{Username: res.Record().Values[0].(string), Mutuals: res.Record().Values[1].(int64)}
			recs = append(recs, rec)
		}
		return recs, nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": username, "recommendations": data})
}
