DELIMITER //

DROP PROCEDURE IF EXISTS create_test_users //
CREATE PROCEDURE create_test_users()
BEGIN

  -- Hash the passwords
  DECLARE all_pw VARCHAR(255);
  SET all_pw = '$2a$10$b.WsKp9GR.8pcdQjxMggGeCtTL7nvuc1oW2LfZu0FrM5SLv3dhkge'; -- pw123

  -- End User
  INSERT INTO users (username, email, password, role)
  VALUES ('end_user', 'end_user@example.com', all_pw, 'end_user');

  -- Premium User  
  INSERT INTO users (username, email, password, role)
  VALUES ('premium_user', 'premium_user@example.com', all_pw, 'premium_user');

  -- System Admin
  INSERT INTO users (username, email, password, role)
  VALUES ('sys_admin', 'sys_admin@example.com', all_pw, 'sys_admin');

  -- Super Admin
  INSERT INTO users (username, email, password, role)
  VALUES ('super_admin', 'super_admin@example.com', all_pw, 'super_admin');

END //

DELIMITER ;

CALL create_test_users();

-- DELETE FROM users;