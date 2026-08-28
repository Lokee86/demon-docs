package codemaparcana

func (resolver *Resolver) Close() error {
	if resolver == nil {
		return nil
	}
	var firstErr error
	if resolver.client != nil {
		firstErr = resolver.client.Close()
	}
	resolver.historyMu.Lock()
	defer resolver.historyMu.Unlock()
	for directory, client := range resolver.historicalClients {
		if err := client.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(resolver.historicalClients, directory)
	}
	return firstErr
}
