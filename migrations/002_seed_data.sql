-- Seed data for Hotel Booking Microservices

-- 1. Insert Admin User (password: Secret123!)
INSERT INTO users (id, email, password, role) VALUES 
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'admin@example.com', '$2a$10$vI8aWBnW3fID.ZQ4/zo1G.q1lRps.9cGLcZEiGDMVr5yUP1KUOYTa', 'admin')
ON CONFLICT (email) DO NOTHING;

-- 2. Insert Customer User (password: Secret123!)
INSERT INTO users (id, email, password, role) VALUES 
('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'customer@example.com', '$2a$10$vI8aWBnW3fID.ZQ4/zo1G.q1lRps.9cGLcZEiGDMVr5yUP1KUOYTa', 'customer')
ON CONFLICT (email) DO NOTHING;

-- 3. Insert Hotel
INSERT INTO hotels (id, name, description, address) VALUES 
('h0eebc99-9c0b-4ef8-bb6d-6bb9bd380h11', 'Grand Hotel Indonesia', 'Luxury hotel in the heart of Jakarta', 'Jl. MH Thamrin No. 1, Jakarta')
ON CONFLICT DO NOTHING;

-- 4. Insert Room Types
INSERT INTO room_types (id, hotel_id, name, capacity, base_price, amenities) VALUES 
('rt0ebc99-9c0b-4ef8-bb6d-6bb9bd38rt01', 'h0eebc99-9c0b-4ef8-bb6d-6bb9bd380h11', 'Deluxe Room', 2, 1500000, 'WiFi, AC, Breakfast'),
('rt0ebc99-9c0b-4ef8-bb6d-6bb9bd38rt02', 'h0eebc99-9c0b-4ef8-bb6d-6bb9bd380h11', 'Executive Suite', 4, 3500000, 'WiFi, AC, Breakfast, Bathtub, City View')
ON CONFLICT DO NOTHING;

-- 5. Insert Rooms
INSERT INTO rooms (id, room_type_id, number, status) VALUES 
(uuid_generate_v4(), 'rt0ebc99-9c0b-4ef8-bb6d-6bb9bd38rt01', '101', 'available'),
(uuid_generate_v4(), 'rt0ebc99-9c0b-4ef8-bb6d-6bb9bd38rt01', '102', 'available'),
(uuid_generate_v4(), 'rt0ebc99-9c0b-4ef8-bb6d-6bb9bd38rt02', '201', 'available')
ON CONFLICT DO NOTHING;
