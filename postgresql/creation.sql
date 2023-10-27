-- Connessione al database "project"
\c project

-- Creazione della tabella 'employee'
CREATE TABLE IF NOT EXISTS employee (
  employee_id INT NOT NULL,
  name VARCHAR(20) NOT NULL,
  surname VARCHAR(20) NOT NULL,
  password VARCHAR(20) NOT NULL,
  PRIMARY KEY (employee_id)
);

-- Inserimento di un record nella tabella 'student'
INSERT INTO employee (employee_id, name, surname, password)
VALUES 
    ('198675', 'Giacomo', 'Zarba Meli', 'password1'),
    ('198676', 'Alessandro', 'Odri', 'password2'),
    ('198677', 'Riccardo', 'Mariotti', 'password3');