# Ubiquitous Language Glossary
## Hotel Booking Domain

This document defines the ubiquitous language used throughout the project. These terms should be used consistently in code, documentation, and conversations.

### Core Concepts

#### Booking (Pesanan)
An agreement between a User and a Hotel for a stay.
- **Aggregate Root**: The main entity managing the lifecycle of a reservation.
- **Identity**: Uniquely identified by a UUID.

#### Booking Status (Status Pesanan)
The current state of a booking lifecycle.
- **Pending Payment**: Booking created, waiting for payment.
- **Confirmed**: Payment successful, room reserved.
- **Cancelled**: Booking terminated before check-in.
- **Checked In**: Guest has arrived at the hotel.
- **Completed**: Guest has checked out, stay finished.

#### Room Type (Tipe Kamar)
A category of rooms available in a hotel (e.g., Deluxe, Suite).
- **Base Price**: The standard nightly rate for this room type.

#### Guest (Tamu)
The person(s) staying in the room.
- **Guest Count**: Number of people included in the booking.

### Financial Terms

#### Money (Uang)
A value object representing a monetary amount and currency.
- **Currency**: Default is IDR (Indonesian Rupiah).
- **Amount**: The numerical value.

#### Total Price (Total Harga)
The final amount to be paid for the booking.
- **Calculation**: `(Nights * Base Price) + Surcharges - Discounts`.

### Time Concepts

#### Date Range (Rentang Tanggal)
The period of stay.
- **Check-In Date**: The day the stay begins (starts at 14:00 local time).
- **Check-Out Date**: The day the stay ends (ends at 12:00 local time).
- **Nights**: The duration of the stay in nights.

### Services

#### Pricing Service
Domain service responsible for calculating the final price, applying complex rules like seasonal rates, discounts, and surcharges.

### System Mechanisms

#### Soft Delete (Penghapusan Lunak)
A technique where data is not permanently removed from the database but marked as deleted using a timestamp (`deleted_at`).
- **Active**: `deleted_at` is NULL.
- **Deleted**: `deleted_at` has a timestamp value.

#### Auto-Checkout (Checkout Otomatis)
A background process (CronJob) that automatically completes bookings when the checkout date is reached.
- **Schedule**: Runs daily at 10:00 AM.
- **Target**: Bookings with `status=checked_in` and `check_out_date=today`.
