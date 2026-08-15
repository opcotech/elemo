package repository

// NotificationProjection selects bounded fields for notification reads.
// PostgreSQL does not compile Cypher; these types keep Get/List signatures
// consistent with other repositories.
type NotificationProjection struct{}

func NotificationListProjection() NotificationProjection {
	return NotificationProjection{}
}

func NotificationDetailProjection() NotificationProjection {
	return NotificationProjection{}
}
