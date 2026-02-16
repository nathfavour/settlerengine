module github.com/nathfavour/settlerengine/settlerd

go 1.24.0

require (
	github.com/nathfavour/settlerengine/core v0.0.0
	github.com/nathfavour/settlerengine/pkg v0.0.0
)

replace (
	github.com/nathfavour/settlerengine/core => ../../core
	github.com/nathfavour/settlerengine/pkg => ../../pkg
)
