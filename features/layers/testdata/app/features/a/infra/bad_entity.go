package infra

// The port of features/a names no entities/pair, so its infra may not import it.
import _ "example.com/app/entities/pair" // want `a feature's infra imports only the entities its port features/a/port imports`
