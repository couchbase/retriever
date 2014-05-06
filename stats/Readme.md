Stats collection API. 

Example usage
'''
func main() {

    const NUM_CONNECTIONS "num_connections"

    sc := stats.NewStatsCollector("test")
    sc.AddStatKey(NUM_CONNECTIONS, 0) 
    sc.UpdateStat(NUM_CONNECTIONS, 10)
    sc.DecrementStat(NUM_CONNECTIONS)

    ...

    fmt.Printf ("Current connections %v", sc.GetStat(NUM_CONNECTIONS)

    fmt.Printf(" All stats %v", sc.GetAllStat())

}
'''
