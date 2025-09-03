package handlers

import (
	pb "api-gateway/proto/follower"
	utils "api-gateway/utils"
	"context"
	"errors"
	"follower-service/db"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const relF = "FOLLOWS"

type FollowerServer struct {
	pb.UnimplementedFollowerServiceServer
}

func NewFollowerServer() *FollowerServer {
	return &FollowerServer{}
}

func (s *FollowerServer) Follow(ctx context.Context, req *pb.FollowRequest) (*pb.FollowResponse, error) {
	fromUsername, _, err := utils.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	toUsername := req.To

	if fromUsername == toUsername {
		return nil, errors.New("cannot follow yourself")
	}

	session := db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `
		MERGE (a:User {username:$from})
		MERGE (b:User {username:$to})
		MERGE (a)-[:` + relF + `]->(b)`
		_, err := tx.Run(ctx, q, map[string]any{"from": fromUsername, "to": toUsername})
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return &pb.FollowResponse{
		Status: "followed successfully",
	}, nil
}

func (s *FollowerServer) Unfollow(ctx context.Context, req *pb.UnfollowRequest) (*pb.UnfollowResponse, error) {
	fromUsername, _, err := utils.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	toUsername := req.To

	if fromUsername == toUsername {
		return nil, errors.New("cannot unfollow yourself")
	}

	session := db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `
		MATCH (a:User {username:$from})-[r:` + relF + `]->(b:User {username:$to})
		DELETE r
		RETURN COUNT(*)`
		res, err := tx.Run(ctx, q, map[string]any{"from": fromUsername, "to": toUsername})
		if err != nil {
			return nil, err
		}
		if res.Next(ctx) {
			return res.Record().Values[0], nil
		}
		return nil, errors.New("no result")
	})
	if err != nil {
		return nil, err
	}
	return &pb.UnfollowResponse{
		Status: "unfollowed successfully",
	}, nil
}

func (s *FollowerServer) GetFollowing(ctx context.Context, req *pb.GetFollowingRequest) (*pb.GetFollowingResponse, error) {
	log.Printf("Primljen zahtev za GetFollowing za korisnika: %s", req.Username)
	username := req.Username
	session := db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	data, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `MATCH (:User {username:$u})-[:` + relF + `]->(f:User) RETURN f.username AS u ORDER BY u`
		res, err := tx.Run(ctx, q, map[string]any{"u": username})
		if err != nil {
			return nil, err
		}
		usernames := make([]string, 0)
		for res.Next(ctx) {
			usernames = append(usernames, res.Record().Values[0].(string))
		}
		return usernames, nil
	})
	if err != nil {
		return nil, err
	}
	followingUsernames, ok := data.([]string)
	if !ok {
		return nil, errors.New("invalid data format")
	}

	return &pb.GetFollowingResponse{
		Following: followingUsernames,
	}, nil
}

func (s *FollowerServer) GetFollowers(ctx context.Context, req *pb.GetFollowersRequest) (*pb.GetFollowersResponse, error) {
	username := req.Username
	session := db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	data, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `MATCH (f:User)-[:` + relF + `]->(:User {username:$u}) RETURN f.username AS u ORDER BY u`
		res, err := tx.Run(ctx, q, map[string]any{"u": username})
		if err != nil {
			return nil, err
		}
		usernames := make([]string, 0)
		for res.Next(ctx) {
			usernames = append(usernames, res.Record().Values[0].(string))
		}
		return usernames, nil
	})
	if err != nil {
		return nil, err
	}
	followerUsernames, ok := data.([]string)
	if !ok {
		return nil, errors.New("invalid data format")
	}

	return &pb.GetFollowersResponse{
		Followers: followerUsernames,
	}, nil
}

func (s *FollowerServer) Recommend(ctx context.Context, req *pb.RecommendRequest) (*pb.RecommendResponse, error) {
	username, _, err := utils.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	limit := 10

	session := db.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	data, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		q := `
		MATCH (me:User {username:$u})-[:` + relF + `]->(m:User)-[:` + relF + `]->(rec:User)
		WHERE NOT (me)-[:` + relF + `]->(rec) AND me <> rec
		RETURN rec.username AS username, COUNT(DISTINCT m) AS mutuals
		ORDER BY mutuals DESC, username ASC
		LIMIT $limit`
		res, err := tx.Run(ctx, q, map[string]any{"u": username, "limit": limit})
		if err != nil {
			return nil, err
		}
		recs := make([]*pb.RecDTO, 0)
		for res.Next(ctx) {
			rec := &pb.RecDTO{Username: res.Record().Values[0].(string), Mutuals: res.Record().Values[1].(int64)}
			recs = append(recs, rec)
		}
		return recs, nil
	})
	if err != nil {
		return nil, err
	}
	recommendations, ok := data.([]*pb.RecDTO)
	if !ok {
		return nil, errors.New("invalid data format")
	}

	return &pb.RecommendResponse{
		RecommendedUsers: recommendations,
	}, nil
}
