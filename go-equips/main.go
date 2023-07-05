package main

import (
    "os"
    "context"
	"fmt"
    "strings"
	"log"
    "net/http"
    "math/big"
    "github.com/gin-gonic/gin"
    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/core/types"
    "database/sql"
	"github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

type Event struct {
	Owner        string
	TokenAddress string
	TokenID      int64
}

func dropTable(db *sql.DB, tableName string) error {
    query := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
    _, err := db.Exec(query)
    if err != nil {
        return err
    }
    return nil
}

func getItemsByParameter(parameter string) ([]Event, error) {
    db, err := sql.Open("postgres", "postgres://"+os.Getenv("PG_USER")+":"+os.Getenv("PG_PASSWORD")+"@localhost/equips?sslmode=disable")

    if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT owner, token_address, token_id FROM equips WHERE owner = $1", parameter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Event

	for rows.Next() {
		var item Event
        err := rows.Scan(&item.Owner, &item.TokenAddress, &item.TokenID)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func getEventLogs() ([]Event, error) {
    db, err := sql.Open("postgres", "postgres://"+os.Getenv("PG_USER")+":"+os.Getenv("PG_PASSWORD")+"@localhost/equips?sslmode=disable")

    if err != nil {
        return nil, err
    }
    
    defer db.Close()

    rows, err := db.Query("SELECT owner, token_address, token_id FROM equips;")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []Event

    for rows.Next() {
        var item Event
        err := rows.Scan(&item.Owner, &item.TokenAddress, &item.TokenID)
        if err != nil {
            return nil, err
        }

        fmt.Printf("Owner: %s, TokenAddress: %s, TokenID: %d\n", item.Owner, item.TokenAddress, item.TokenID)
        
        items = append(items, item)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return items, nil
}

func handleAddItem(c *gin.Context) {
    db, err := sql.Open("postgres", "postgres://"+os.Getenv("PG_USER")+":"+os.Getenv("PG_PASSWORD")+"@localhost/equips?sslmode=disable")

    if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

    for i := 0; i < 1500; i++ {
            insertSQL := `
            INSERT INTO equips (owner, token_address, token_id)
            VALUES ($1, $2, $3)
        `
        // Execute the insert statement
        _, err = db.Exec(insertSQL, "0xbabe", "0xdeaf", i)
        if err != nil {
            log.Fatal(err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        fmt.Println("Table added to")
	}
    c.JSON(http.StatusOK, gin.H{"status": "Table added to"})
}

func handleInitTable(c *gin.Context) {
    db, err := sql.Open("postgres", "postgres://"+os.Getenv("PG_USER")+":"+os.Getenv("PG_PASSWORD")+"@localhost/equips?sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
    // Create a table
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS equips (
        id SERIAL PRIMARY KEY,
        owner VARCHAR(42),
        token_address VARCHAR(42),
        token_id INT
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return 
	}

	fmt.Println("Table created successfully")
    c.JSON(http.StatusOK, gin.H{"status": "Table created successfully"})
}

func handleParamLookUp(c *gin.Context){
    address := c.Param("address")
    items, err := getItemsByParameter(address)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"items": items})
}

func handleEventLogs(c *gin.Context) {
    logs, err := getEventLogs()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func handleDropTable(c *gin.Context){
    db, err := sql.Open("postgres", "postgres://"+os.Getenv("PG_USER")+":"+os.Getenv("PG_PASSWORD")+"@localhost/equips?sslmode=disable")

    // Drop the table
    err = dropTable(db, "equips")
    if err != nil {
        fmt.Println("Error dropping table:", err)
        return
    }

    fmt.Println("Table dropped successfully!")
}

func listen() {
    db, err := sql.Open("postgres", "postgres://"+os.Getenv("PG_USER")+":"+os.Getenv("PG_PASSWORD")+"@localhost/equips?sslmode=disable")

    client, err := ethclient.Dial("wss://polygon-mumbai.g.alchemy.com/v2/"+os.Getenv("ALCHEMY_RPC_WSS"))
    if err != nil {
        log.Println("err")
        log.Fatal(err)
    }

    eventName := `Equip`
    contractABI := `
    [
        {
            "inputs": [
                {
                    "internalType": "address[]",
                    "name": "token_addresses",
                    "type": "address[]"
                },
                {
                    "internalType": "uint256[]",
                    "name": "token_ids",
                    "type": "uint256[]"
                },
                {
                    "internalType": "uint256",
                    "name": "salt",
                    "type": "uint256"
                }
            ],
            "name": "add",
            "outputs": [],
            "stateMutability": "nonpayable",
            "type": "function"
        },
        {
            "inputs": [],
            "stateMutability": "nonpayable",
            "type": "constructor"
        },
        {
            "anonymous": false,
            "inputs": [
                {
                    "indexed": true,
                    "internalType": "address",
                    "name": "owner",
                    "type": "address"
                },
                {
                    "indexed": true,
                    "internalType": "address",
                    "name": "token_address",
                    "type": "address"
                },
                {
                    "indexed": true,
                    "internalType": "uint256",
                    "name": "token_id",
                    "type": "uint256"
                },
                {
                    "indexed": false,
                    "internalType": "uint256",
                    "name": "id",
                    "type": "uint256"
                },
                {
                    "indexed": false,
                    "internalType": "uint256",
                    "name": "salt",
                    "type": "uint256"
                }
            ],
            "name": "Equip",
            "type": "event"
        }
    ]
	`
    parsedABI, err := abi.JSON(strings.NewReader(contractABI))
    contractAddress := common.HexToAddress("0xb26081B54ae4D9F441025DeE71597E82c82077E1")
    query := ethereum.FilterQuery{
        Addresses: []common.Address{contractAddress},
        Topics:    [][]common.Hash{{parsedABI.Events[eventName].ID}},
    }

    logs := make(chan types.Log)
    sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
    if err != nil {
        log.Println("err")
        log.Fatal(err)
    }

    for {
        select {
            case err := <-sub.Err():
                log.Fatal(err)
            case eventLog := <-logs:
                insertSQL := `
                    INSERT INTO equips (owner, token_address, token_id)
                    VALUES ($1, $2, $3)
                `
                // Execute the insert statement
                _, err = db.Exec(
                    insertSQL, 
                    common.BytesToAddress(eventLog.Topics[1].Bytes()).Hex(),
                    common.BytesToAddress(eventLog.Topics[2].Bytes()).Hex(),
                    new(big.Int).SetBytes(eventLog.Topics[3].Bytes()).Int64(),
                )
                fmt.Println("Table added to")
        }
    }
}

func main() {

    err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

    go func() {
        listen()
    }()

    r := gin.Default()

    // *~*~*~*~ for testing
    r.GET("/init", handleInitTable)
    r.GET("/add", handleAddItem)
    r.GET("/drop", handleDropTable)

    // *~*~*~*~ for api
    r.GET("/all", handleEventLogs)
    r.GET("/lookup/:address", handleParamLookUp)

    // @^@^@ listening
    r.Run(":7077")
}