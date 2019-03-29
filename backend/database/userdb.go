package database

type UserDB struct {
	adb *AdminDB

	user string
}

func NewUserDB(adb *AdminDB, user string) *UserDB {
	return &UserDB{
		adb:  adb,
		user: user,
	}
}

// AdminDB returns the admin database
func (db *UserDB) AdminDB() *AdminDB {
	return db.adb
}

func (db *UserDB) ID() string {
	return db.user
}

func (db *UserDB) CreateUser(u *User) error {

	// Only create the user if I have the users:create scope
	return createUser(db.adb, u, `SELECT 1 FROM group_scopes WHERE scope='users:create' AND (
			groupid IN (
				SELECT groupid FROM group_members WHERE username=?
			) OR groupid IN (?, 'public', 'users')
		) LIMIT 1;`, db.user, db.user)
}

func (db *UserDB) ReadUser(name string) (*User, error) {
	// A user can be read if:
	//	the user's public_access is >= 100 (read access by public),
	//	the user's user_access >=100
	//	the user is member of a group which gives it users:read scope
	//	the user to be read is itself, and the user has user:read scope

	if name != db.user {
		return readUser(db.adb, name, `SELECT * FROM groups WHERE id=? AND owner=id 
		AND (
				(public_access >= 100 OR user_access >=100)
			OR EXISTS 
				(SELECT 1 FROM group_scopes WHERE scope='users:read' AND (
						groupid IN (?, 'public', 'users') 
					OR 
						groupid IN (SELECT groupid FROM group_members WHERE username=?)
					)
				)
		)
		LIMIT 1;`, name, db.user, db.user)

	}

	return readUser(db.adb, name, `SELECT * FROM groups WHERE id=? AND owner=id 
		AND (
				EXISTS (SELECT 1 from group_scopes WHERE scope='user:read' and groupid=?)
			OR
				(public_access >= 100 OR user_access >=100)
			OR EXISTS 
				(SELECT 1 FROM group_scopes WHERE scope='users:read' AND (
						groupid IN (?, 'public', 'users') 
					OR 
						groupid IN (SELECT groupid FROM group_members WHERE username=?)
					)
				)
		)
		LIMIT 1;`, name, db.user, db.user, db.user)

}

// UpdateUser updates the given portions of a user
func (db *UserDB) UpdateUser(u *User) error {
	if u.ID == db.user {
		return updateUser(db.adb, u, `SELECT DISTINCT(scope) FROM group_scopes WHERE 
			(
				(
					scope LIKE 'users:edit%'
				AND
					(groupid IN (SELECT groupid FROM group_members WHERE username=?) OR groupid IN (?, 'public', 'users'))
				)
			OR
				(
					groupid=? 
				AND 
					scope LIKE 'user:edit%'
				)
			);`, db.user, db.user, db.user)
	}
	return updateUser(db.adb, u, `SELECT DISTINCT(scope) FROM group_scopes WHERE 
		(
			(
				scope LIKE 'users:edit%'
			AND
				(groupid IN (SELECT groupid FROM group_members WHERE username=?) OR groupid IN (?, 'public', 'users'))
			)
		);`, db.user, db.user)
}

func (db *UserDB) DelUser(name string) error {
	// A user can be deleted if:
	//	the user is member of a group which gives it users:delete scope
	//	the user to be read is itself, and the user has user:delete scope

	if db.user != name {
		return delUser(db.adb, name, `DELETE FROM users WHERE name=? AND EXISTS (
			SELECT 1 FROM group_scopes WHERE scope='users:delete' AND (
				groupid IN (?, 'public', 'users') 
			OR 
				groupid IN (SELECT groupid FROM group_members WHERE username=?)
			) LIMIT 1
		);`, name, db.user, db.user)
	}

	// The user has same name, add check for user:delete
	return delUser(db.adb, name, `DELETE FROM users WHERE name=? AND EXISTS (
			SELECT 1 FROM group_scopes WHERE scope='user:delete' AND groupid=?
				OR (
					scope='users:delete' AND (
							groupid IN (?, 'public', 'users') 
						OR 
							groupid IN (SELECT groupid FROM group_members WHERE username=?)
						)
			) LIMIT 1
		);`, name, db.user, db.user, db.user)
}
