package main

import "fmt"
import "bufio"
import "net"
import "time"
import "os"
import "flag"
import "gopkg.in/redis.v4"

func resolveTarget(target string) (ip net.IP, err error) {
	ips, err := net.LookupIP(target)
	if err != nil {
		return nil, err
	}

	for _, i := range ips {
		if i.To4() != nil {
			ip = i
			break
		}
	}

	return ip, nil
}

func doHealthCheck(ip net.IP) (respTime int, err error) {
	fmt.Printf("Pinging service... ")

	t1 := time.Now()

	conn, err := net.DialTimeout("tcp", ip.String() + ":80", 1 * time.Second)
	if err != nil {
		return 0, err
	}

	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return 0, err
	}

	t2 := time.Now()
	respTime = int((t2.Sub(t1)) / time.Microsecond)

	fmt.Printf("service returned %v bytes in %v us\n", len(status), respTime)
	return respTime, nil
}

func storeInRedis(redisClient *redis.Client, target string, respTime interface{}) (err error) {
	redisClient.Set("response_times:"+target, respTime, 0)
	return nil
}

func main() {

	target := flag.String("target", "", "Target host for health check")
	interval := flag.Int("interval", 1, "Interval in seconds for health check")
	redisHost := flag.String("redis", "", "Redis server for stat storing")

	flag.Parse()

	redisClient := redis.NewClient(&redis.Options{
		Addr: *redisHost + ":6379",
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	ip, err := resolveTarget(*target)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	for {
		time.Sleep(time.Duration(*interval) * time.Second)

		respTime, err := doHealthCheck(ip)
		if err != nil {
			fmt.Printf("Error during health check\n")
			storeInRedis(redisClient, *target, "error")
			continue
		}

		storeInRedis(redisClient, *target, respTime)
	}
}
