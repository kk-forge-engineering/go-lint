package db

// Root infra stays generic, so it imports no entity.
import _ "example.com/app/entities/user" // want `infra may not import entities \(example\.com/app/entities/user\)`
