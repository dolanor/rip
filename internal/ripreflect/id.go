package ripreflect

// GetID gets the id of an entity.
// Make it use reflect for getting fields of `rip:"id"`.
func GetID(entity any) (string, error) {
	idVal, _, err := FindEntityID(entity)
	if err != nil {
		return "", err
	}

	id := idVal.String()

	return id, nil
}

// SetID set the ID of the entity.
// the entity must be passed by reference.
func SetID(entity any, uuid string) error {
	idVal, _, err := FindEntityID(entity)
	if err != nil {
		return err
	}

	idVal.SetString(uuid)

	return nil
}
