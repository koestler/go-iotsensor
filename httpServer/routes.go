package httpServer

var wsRoutes = WsRoutes{
	WsRoute{
		"ws-test",
		"/ws/v0/RoundedValues",
		HandleWsRoundedValues,
	},
}

var httpRoutes = HttpRoutes{
	HttpRoute{
		"DeviceIndex",
		"GET",
		"/api/v0/device/",
		HandleDeviceIndex,
	},
	HttpRoute{
		"FrontendConfig",
		"GET",
		"/api/v0/FrontendConfig",
		HandleFrontendConfig,
	},
	HttpRoute{
		"RoundedValues",
		"GET",
		"/api/v0/device/{DeviceId:[a-zA-Z0-9\\-]{1,32}}/RoundedValues",
		HandleDeviceGetRoundedValues,
	},
	HttpRoute{
		"DevicePictureThumb",
		"GET",
		"/api/v0/device/{DeviceId:[a-zA-Z0-9\\-]{1,32}}/Picture/Thumb",
		HandleDeviceGetPictureThumb,
	},
	HttpRoute{
		"DevicePictureRaw",
		"GET",
		"/api/v0/device/{DeviceId:[a-zA-Z0-9\\-]{1,32}}/Picture/Raw",
		HandleDeviceGetPictureRaw,
	},
	HttpRoute{
		"DeviceRoundedValuesWebSocket",
		"GET",
		"/api/v0/ws/RoundedValues",
		HandleWsRoundedValues,
	},
	HttpRoute{
		"ApiIndex",
		"GET",
		"/api{Path:.*}",
		HandleApiNotFound,
	},
	HttpRoute{
		"Assets",
		"GET",
		"/{Path:.+}",
		HandleAssetsGet,
	},
}
