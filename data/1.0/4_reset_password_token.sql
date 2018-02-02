CREATE TABLE reset_password_token (
   id UUID PRIMARY KEY ,
   user_id FOREIGN KEY REFERENCES users(id),
   token VARCHAR(255),
   is_used BOOLEAN,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)