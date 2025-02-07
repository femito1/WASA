package api

// Close shuts down any background routines or resources held by the API router.
func (rt *_router) Close() error {
	rt.baseLogger.Info("shutting down API router")
	return nil
}
