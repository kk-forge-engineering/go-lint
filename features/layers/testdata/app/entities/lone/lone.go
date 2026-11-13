// The analyzer reports this entity because it counts only features/b as a consumer.
// An entrypoint, an infra adapter, a test, a package main generator, a file the go tool
// ignores in features/e and a nested Go module under features/ also import it,
// and the analyzer does not count them.
package lone // want `only features/b imports entities/lone: move it into features/b, since an entity is what two or more of them share`
