CREATE TABLE reset_password_token (
   id UUID PRIMARY KEY ,
   user_id FOREIGN KEY REFERENCES users(id),
   token UUID,
   is_expired BOOLEAN,
   is_used BOOLEAN,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)