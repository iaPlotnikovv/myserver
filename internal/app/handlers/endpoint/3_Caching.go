package endpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	dbase "github.com/iaPlotnikovv/myserver/internal/app/init/database"
	rc "github.com/iaPlotnikovv/myserver/internal/app/init/redis_cache"
)

func Caching(zapis string) (*JsonResponse, error) {

	ctx := context.Background()

	client := rc.InitRed()

	cachedComms, err := client.Get(ctx, zapis).Bytes()

	response := JsonResponse{}

	if err != nil {

		dbComms, err := FetchFromDB(zapis)

		if err != nil {
			return nil, err
		}

		cachedComms, err = json.Marshal(dbComms)

		if err != nil {
			return nil, err
		}

		err = client.Set(ctx, zapis, cachedComms, 2*time.Minute).Err()

		response = JsonResponse{Type: "PostgreSQL", Data: dbComms}

		PrintMessage("FROM PostgreSQL...")

		return &response, err
	}

	comms := []info_js{}

	err = json.Unmarshal(cachedComms, &comms)

	if err != nil {
		return nil, err
	}

	response = JsonResponse{Type: "Redis Cache", Data: comms}

	PrintMessage("FROM Redis Cache...")

	return &response, nil
}

func FetchFromDB(zapis string) ([]info_js, error) {

	db := dbase.Init()

	queryString := fmt.Sprintf("SELECT id, comment FROM test WHERE comment LIKE '%s'", zapis)

	rows, err := db.Query(queryString)

	if err != nil {
		return nil, err
	}

	var info []info_js

	for rows.Next() {

		p := info_js{}

		err = rows.Scan(&p.ID, &p.Comment)

		info = append(info, p)

		if err != nil {
			return nil, err
		}

	}

	return info, nil
}
