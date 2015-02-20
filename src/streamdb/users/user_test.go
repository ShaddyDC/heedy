package users

import(
    "testing"
    "reflect"
    )

func TestCreateUser(t *testing.T) {
    _, err := testdb.CreateUser("TestCreateUser_name", "TestCreateUser_email", "TestCreateUser_pass")
    if err != nil {
        t.Errorf("Cannot create user %v", err)
        return
    }

    _, err = testdb.CreateUser("TestCreateUser_name", "TestCreateUser_email2", "TestCreateUser_pass2")
    if err == nil {
        t.Errorf("Created duplicate user name %v", err)
        return
    }

    _, err = testdb.CreateUser("TestCreateUser_name2", "TestCreateUser_email", "TestCreateUser_pass2")
    if err == nil {
        t.Errorf("Created duplicate email %v", err)
        return
    }
}

func TestReadUserByEmail(t *testing.T){
    // test failures on non existance
    usr, err := testdb.ReadUserByEmail("doesnotexist   because spaces")

    if usr != nil {
        t.Errorf("Selected user that does not exist by email")
    }

    if err == nil {
        t.Errorf("no error returned, expected non nil on failing case")
    }

    // setup for reading
    _, err = testdb.CreateUser("TestReadUserByEmail_name", "TestReadUserByEmail_email", "TestReadUserByEmail_pass")
    if err != nil {
        t.Errorf("Could not create user for test reading... %v", err)
        return
    }

    usr, err = testdb.ReadUserByEmail("TestReadUserByEmail_email")

    if usr == nil {
        t.Errorf("did not get a user by email")
    }

    if err != nil {
        t.Errorf("got an error when trying to get a user that should exist %v", err)
    }
}


func TestReadUserByName(t *testing.T){
    // test failures on non existance
    usr, err := testdb.ReadUserByName("")

    if usr != nil {
        t.Errorf("Selected user that does not exist by name")
    }

    if err == nil {
        t.Errorf("no error returned, expected non nil on failing case")
    }

    // setup for reading
    _, err = testdb.CreateUser("TestReadUserByName_name", "TestReadUserByName_email", "TestReadUserByName_pass")
    if err != nil {
        t.Errorf("Could not create user for test reading... %v", err)
        return
    }

    usr, err = testdb.ReadUserByName("TestReadUserByName_name")

    if usr == nil {
        t.Errorf("did not get a user by name")
    }

    if err != nil {
        t.Errorf("got an error when trying to get a user that should exist %v", err)
    }
}

func TestValidateUser(t *testing.T){
    name, email, pass := "TestValidateUser_name", "TestValidateUser_email", "TestValidateUser_pass"

    _, err := testdb.CreateUser(name, email, pass)
    if err != nil {
        t.Errorf("Cannot create user %v", err)
        return
    }

    validated, _ := testdb.ValidateUser(name, pass)
    if ! validated  {
        t.Errorf("could not validate a user with username and pass")
    }

    validated, _ = testdb.ValidateUser(email, pass)
    if ! validated {
        t.Errorf("could not validate a user with email and pass")
    }

    validated, _ = testdb.ValidateUser(email, email)
    if validated {
        t.Errorf("Validated an incorrect user")
    }
}


func TestReadUserById(t *testing.T){
    // test failures on non existance
    usr, err := testdb.ReadUserById(-1)

    if usr != nil {
        t.Errorf("Selected user that does not exist by name")
    }

    if err == nil {
        t.Errorf("no error returned, expected non nil on failing case")
    }

    // setup for reading
    id, err := testdb.CreateUser("ReadUserById_name", "ReadUserById_email", "ReadUserById_pass")
    if err != nil {
        t.Errorf("Could not create user for test reading... %v", err)
        return
    }

    usr, err = testdb.ReadUserById(id)

    if usr == nil {
        t.Errorf("did not get a user by id")
    }

    if err != nil {
        t.Errorf("got an error when trying to get a user that should exist %v", err)
    }
}


func TestUpdateUser(t *testing.T){
    // setup for reading
    id, err := testdb.CreateUser("TestUpdateUser_name", "TestUpdateUser_email", "TestUpdateUser_pass")
    if err != nil {
        t.Errorf("Could not create user for test reading... %v", err)
        return
    }

    usr, err := testdb.ReadUserById(id)

    if usr == nil {
        t.Errorf("did not get a user by id")
        return
    }

    if err != nil {
        t.Errorf("got an error when trying to get a user that should exist %v", err)
        return
    }

    usr.Name = "Hello"
    usr.Email = "hello@example.com"
    usr.Admin = true
    usr.Phone = "(303) 303-0000" //Non-legal phone number, don't worry
    usr.UploadLimit_Items = 1
    usr.ProcessingLimit_S = 1
    usr.StorageLimit_Gb = 1

    err = testdb.UpdateUser(usr)

    if err != nil {
        t.Errorf("Could not update user %v", err)
    }

    usr2, err := testdb.ReadUserById(id)

    usr.ModifyTime = usr2.ModifyTime // have to do this because we update it each time!
    if err != nil {
        t.Errorf("got an error when trying to get a user that should exist %v", err)
        return
    }

    if ! reflect.DeepEqual(usr, usr2) {
        t.Errorf("The original and updated objects don't match orig: %v updated %v", usr, usr2)
    }
}



func TestDeleteUser(t *testing.T) {
    id, err := testdb.CreateUser("a", "b", "c")

    if nil != err {
        t.Errorf("Cannot create user to test delete")
        return
    }

    err = testdb.DeleteUser(id)

    if nil != err {
        t.Errorf("Error when attempted delete %v", err)
        return
    }

    user, err := testdb.ReadUserById(id)

    if err == nil {
        t.Errorf("The user with ID %v should have errored out, but it did not", id)
        return
    }
    if user != nil {
        t.Errorf("Expected nil, but we got back a user meaning the delete failed %v", user)
    }
}
