
DELIMITER //

DROP PROCEDURE IF EXISTS create_test_users //
CREATE PROCEDURE create_test_users()
BEGIN

  -- End User
  INSERT INTO users (username, email, password, role)
  VALUES ('end_user', 'end_user@example.com', 'password123', 'end_user');

  -- Premium User  
  INSERT INTO users (username, email, password, role)
  VALUES ('premium_user', 'premium_user@example.com', 'password123', 'premium_user');

  -- System Admin
  INSERT INTO users (username, email, password, role)
  VALUES ('sys_admin', 'sys_admin@example.com', 'password123', 'sys_admin');

  -- Super Admin
  INSERT INTO users (username, email, password, role)
  VALUES ('super_admin', 'super_admin@example.com', 'password123', 'super_admin');

END //

DELIMITER ;

CALL create_test_users();