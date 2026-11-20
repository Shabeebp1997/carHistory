# CarHistory — Complete API Reference

**Base URL:** `http://localhost:8080/api/v1`  
**Auth:** `Authorization: Bearer <access_token>` on all protected routes  
**Content-Type:** `application/json` unless noted as `multipart/form-data`  
**Dates:** ISO 8601 string — `"2026-10-24"` or `"2026-10-24T18:00:00Z"`

> **Postman tip:** Set a collection variable `{{base}}` = `http://localhost:8080/api/v1` and `{{token}}` = your access token.  
> **Windows PowerShell:** use `curl.exe` instead of `curl`.

---

## 1. Authentication

### POST `/auth/register` — Create account
No auth required. Pass `"is_admin": true` **with an admin Bearer token** to create an admin account.

```bash
curl -X POST {{base}}/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Shabeeb Mohammed",
    "email": "shabeeb@email.com",
    "mobile_number": "+919074058044",
    "password": "MyPass@123",
    "is_admin": false
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | ✅ | 2–255 chars |
| `email` | string | ⚠️ | Required if no `mobile_number` |
| `mobile_number` | string | ⚠️ | Required if no `email` |
| `password` | string | ✅ | Min 8 chars |
| `is_admin` | bool | ❌ | Only honoured when caller provides a valid admin token |

**Response `201`:**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "user": { "id": "uuid", "name": "Shabeeb Mohammed", "email": "shabeeb@email.com", "is_admin": false }
}
```

---

### POST `/auth/login` — Sign in

```bash
curl -X POST {{base}}/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "shabeeb@email.com",
    "password": "MyPass@123"
  }'
```

Login with phone instead:
```json
{ "phone": "+919074058044", "password": "MyPass@123" }
```

| Field | Type | Notes |
|---|---|---|
| `email` | string | Use email OR phone |
| `phone` | string | Use email OR phone |
| `password` | string | Required |

---

### POST `/auth/refresh` — Refresh tokens

```bash
curl -X POST {{base}}/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{ "refresh_token": "eyJ..." }'
```

---

### POST `/auth/logout`

```bash
curl -X POST {{base}}/auth/logout \
  -H "Content-Type: application/json" \
  -d '{ "refresh_token": "eyJ..." }'
```

---

### POST `/auth/forgot-password`

```bash
curl -X POST {{base}}/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{ "email": "shabeeb@email.com" }'
```

---

### POST `/auth/reset-password`

```bash
curl -X POST {{base}}/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{ "token": "<reset_token_from_email>", "password": "NewPass@456" }'
```

---

## 2. User Profile

### GET `/users/me` — Get my profile

```bash
curl {{base}}/users/me -H "Authorization: Bearer {{token}}"
```

---

### PUT `/users/me` — Update profile

```bash
curl -X PUT {{base}}/users/me \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Shabeeb Mohammed",
    "mobile_number": "+919074058044",
    "timezone": "Asia/Kolkata",
    "email_notifications_enabled": true,
    "sms_notifications_enabled": false
  }'
```

| Field | Type | Notes |
|---|---|---|
| `name` | string | 2–255 chars |
| `mobile_number` | string | Optional |
| `timezone` | string | IANA timezone e.g. `"Asia/Kolkata"` |
| `email_notifications_enabled` | bool | Optional |
| `sms_notifications_enabled` | bool | Optional |

---

### PUT `/users/me/password` — Change password

```bash
curl -X PUT {{base}}/users/me/password \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "MyPass@123",
    "new_password": "NewPass@456"
  }'
```

---

### POST `/users/me/avatar` — Upload profile picture
`multipart/form-data`. Allowed types: jpeg, png, gif, webp.

```bash
curl -X POST {{base}}/users/me/avatar \
  -H "Authorization: Bearer {{token}}" \
  -F "file=@/path/to/photo.jpg"
```

---

## 3. Vehicles

### GET `/vehicles` — List my vehicles

```bash
curl {{base}}/vehicles -H "Authorization: Bearer {{token}}"
```

---

### GET `/vehicles/search` — Search vehicles
Find by registration plate, owner email, or owner phone.

```bash
# By registration plate
curl "{{base}}/vehicles/search?registration=KL10AR8764" \
  -H "Authorization: Bearer {{token}}"

# By owner email
curl "{{base}}/vehicles/search?email=shabeeb@email.com" \
  -H "Authorization: Bearer {{token}}"

# By owner phone
curl "{{base}}/vehicles/search?phone=%2B919074058044" \
  -H "Authorization: Bearer {{token}}"
```

| Query param | Notes |
|---|---|
| `registration` | Partial match on plate number |
| `email` | Exact owner email |
| `phone` | Exact owner phone |

At least one param is required.

---

### POST `/vehicles` — Add a vehicle

```bash
curl -X POST {{base}}/vehicles \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "vehicle_type_id": "UUID_FROM_vehicle-types",
    "brand": "Toyota",
    "model": "Camry",
    "variant": "Hybrid XLE 2.5L",
    "year": 2022,
    "registration_number": "KL10AR8764",
    "chassis_number": "MA3FW61S000123456",
    "engine_number": "2AZFXE98765",
    "fuel_type": "hybrid",
    "colour": "Crimson Red",
    "rc_owner_name": "Shabeeb Mohammed",
    "rc_owner_contact": "+919074058044",
    "current_odometer": 0,
    "purchase_price": 3200000.00,
    "purchase_date": "2022-06-01",
    "notes": "Company car"
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `vehicle_type_id` | UUID | ✅ | From `GET /vehicle-types` |
| `brand` | string | ✅ | e.g. `"Toyota"` |
| `model` | string | ✅ | e.g. `"Camry"` |
| `variant` | string | ❌ | e.g. `"Hybrid XLE"` |
| `year` | int | ✅ | Manufacturing year |
| `registration_number` | string | ✅ | Plate — e.g. `"KL10AR8764"` |
| `chassis_number` | string | ❌ | VIN/chassis |
| `engine_number` | string | ❌ | Engine serial |
| `fuel_type` | string | ❌ | `petrol` `diesel` `electric` `cng` `hybrid` |
| `colour` | string | ❌ | Body colour |
| `rc_owner_name` | string | ❌ | Name on RC book |
| `rc_owner_contact` | string | ❌ | RC owner phone |
| `current_odometer` | int | ❌ | km at time of adding |
| `purchase_price` | float | ❌ | In local currency |
| `purchase_date` | date | ❌ | `"YYYY-MM-DD"` |
| `notes` | string | ❌ | Free text |

---

### GET `/vehicles/:vehicleId` — Get vehicle details

```bash
curl {{base}}/vehicles/VEHICLE_ID -H "Authorization: Bearer {{token}}"
```

---

### PUT `/vehicles/:vehicleId` — Update vehicle
Requires `owner` or `editor` role. Same fields as create.

```bash
curl -X PUT {{base}}/vehicles/VEHICLE_ID \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{ "current_odometer": 45230, "colour": "Pearl White" }'
```

---

### DELETE `/vehicles/:vehicleId` — Archive vehicle
Requires `owner` role.

```bash
curl -X DELETE {{base}}/vehicles/VEHICLE_ID \
  -H "Authorization: Bearer {{token}}"
```

---

### GET `/vehicle-types` — List vehicle types

```bash
curl {{base}}/vehicle-types -H "Authorization: Bearer {{token}}"
```

Returns: `[{ "id": "uuid", "name": "Sedan" }, ...]`

---

## 4. Vehicle Members & Invitations

### GET `/vehicles/:vehicleId/members`

```bash
curl {{base}}/vehicles/VEHICLE_ID/members -H "Authorization: Bearer {{token}}"
```

---

### POST `/vehicles/:vehicleId/invitations` — Invite a user (owner only)

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/invitations \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "friend@email.com",
    "role": "viewer"
  }'
```

| Field | Type | Notes |
|---|---|---|
| `email` | string | Invitee email |
| `role` | string | `owner` `editor` `viewer` |

---

### POST `/invitations/:token/accept`

```bash
curl -X POST {{base}}/invitations/TOKEN/accept \
  -H "Authorization: Bearer {{token}}"
```

---

### PUT `/vehicles/:vehicleId/members/:userId/role` — Change role (owner only)

```bash
curl -X PUT {{base}}/vehicles/VEHICLE_ID/members/USER_ID/role \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{ "role": "editor" }'
```

---

### DELETE `/vehicles/:vehicleId/members/:userId` — Remove member (owner only)

```bash
curl -X DELETE {{base}}/vehicles/VEHICLE_ID/members/USER_ID \
  -H "Authorization: Bearer {{token}}"
```

---

## 5. Maintenance Records

### GET `/vehicles/:vehicleId/maintenance`

```bash
curl {{base}}/vehicles/VEHICLE_ID/maintenance -H "Authorization: Bearer {{token}}"
```

---

### POST `/vehicles/:vehicleId/maintenance`

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/maintenance \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "maintenance_type_id": "UUID_FROM_maintenance-types",
    "date": "2026-10-20",
    "odometer_reading": 45000,
    "description": "Full synthetic oil change + filter",
    "cost": 1850.00,
    "service_provider": "Toyota Authorised — Edappally",
    "next_due_date": "2027-04-20",
    "next_due_odometer": 50000,
    "notes": "Used Mobil 1 5W-30"
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `maintenance_type_id` | UUID | ✅ | From `GET /maintenance-types` |
| `date` | date | ✅ | Service date |
| `odometer_reading` | int | ❌ | km at service |
| `description` | string | ❌ | What was done |
| `cost` | float | ❌ | Amount paid |
| `service_provider` | string | ❌ | Workshop name |
| `next_due_date` | date | ❌ | Next scheduled date |
| `next_due_odometer` | int | ❌ | Next km trigger |
| `notes` | string | ❌ | Extra notes |

---

### PUT `/vehicles/:vehicleId/maintenance/:id` — Update (owner/editor)
### DELETE `/vehicles/:vehicleId/maintenance/:id` — Delete (owner only)

---

## 6. Service Records

### POST `/vehicles/:vehicleId/services`

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/services \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2026-09-10",
    "odometer_reading": 44000,
    "service_center": "Honda Authorised — MG Road",
    "description": "Full periodic service 40k",
    "cost": 4500.00,
    "next_service_date": "2027-03-10",
    "next_service_odometer": 50000,
    "notes": "Replaced air filter and spark plugs"
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `date` | date | ✅ | |
| `odometer_reading` | int | ❌ | |
| `service_center` | string | ❌ | |
| `description` | string | ❌ | |
| `cost` | float | ❌ | |
| `next_service_date` | date | ❌ | |
| `next_service_odometer` | int | ❌ | |
| `notes` | string | ❌ | |

`GET` / `PUT` / `DELETE` follow the same pattern as maintenance.

---

## 7. Insurance

### POST `/vehicles/:vehicleId/insurance`

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/insurance \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "company": "HDFC ERGO General Insurance",
    "policy_number": "2311200217877800",
    "type": "Comprehensive",
    "start_date": "2026-01-01",
    "expiry_date": "2027-01-01",
    "premium": 14500.00,
    "coverage_amount": 3200000.00,
    "agent_name": "Rahul Nair",
    "agent_contact": "+919876543210",
    "notes": "Zero depreciation add-on included"
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `company` | string | ✅ | Insurer name |
| `policy_number` | string | ✅ | Policy ID |
| `type` | string | ❌ | e.g. `"Comprehensive"` `"Third Party"` |
| `start_date` | date | ✅ | |
| `expiry_date` | date | ✅ | |
| `premium` | float | ❌ | Annual premium |
| `coverage_amount` | float | ❌ | Sum insured |
| `agent_name` | string | ❌ | |
| `agent_contact` | string | ❌ | |
| `notes` | string | ❌ | |

---

## 8. PUC / Emission Certificate

### POST `/vehicles/:vehicleId/puc`

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/puc \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "certificate_number": "KL10-PUC-2026-88451",
    "test_date": "2026-10-01",
    "expiry_date": "2026-12-01",
    "test_center": "GreenAir Emissions — Kaloor",
    "notes": ""
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `certificate_number` | string | ✅ | |
| `test_date` | date | ✅ | |
| `expiry_date` | date | ✅ | |
| `test_center` | string | ❌ | |
| `notes` | string | ❌ | |

---

## 9. Tyres

### POST `/vehicles/:vehicleId/tyres`

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/tyres \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "brand": "MRF ZLX",
    "size": "215/55 R17",
    "type": "All Season",
    "quantity": 4,
    "installed_date": "2025-06-15",
    "odometer_reading": 35000,
    "cost": 12400.00,
    "notes": "Set of 4, free rotation included"
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `brand` | string | ✅ | e.g. `"MRF ZLX"` |
| `size` | string | ❌ | e.g. `"215/55 R17"` |
| `type` | string | ❌ | e.g. `"All Season"` `"Performance"` |
| `quantity` | int | ❌ | Number of tyres |
| `installed_date` | date | ❌ | |
| `odometer_reading` | int | ❌ | km at install |
| `cost` | float | ❌ | Total cost |
| `notes` | string | ❌ | |

---

## 10. Battery

### POST `/vehicles/:vehicleId/batteries`

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/batteries \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "brand": "Amaron Pro",
    "model": "AM-PRO-65B24LS",
    "capacity_ah": 65,
    "installed_date": "2025-03-10",
    "odometer_reading": 38000,
    "warranty_months": 36,
    "cost": 5200.00,
    "notes": "Lead-acid 12V"
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `brand` | string | ✅ | |
| `model` | string | ❌ | |
| `capacity_ah` | int | ❌ | Ampere-hours |
| `installed_date` | date | ❌ | |
| `odometer_reading` | int | ❌ | |
| `warranty_months` | int | ❌ | |
| `cost` | float | ❌ | |
| `notes` | string | ❌ | |

---

## 11. Odometer Readings

### POST `/vehicles/:vehicleId/odometer`

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/odometer \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "reading": 45230,
    "date": "2026-10-24",
    "notes": "Monthly reading"
  }'
```

### GET `/vehicles/:vehicleId/odometer` — Reading history

---

## 12. Fuel & Charging Logs

### GET `/vehicles/:vehicleId/fuel`

```bash
curl {{base}}/vehicles/VEHICLE_ID/fuel -H "Authorization: Bearer {{token}}"
```

---

### GET `/vehicles/:vehicleId/fuel/summary`

```bash
curl {{base}}/vehicles/VEHICLE_ID/fuel/summary -H "Authorization: Bearer {{token}}"
```

Returns total spent, total fills, total litres, and breakdown by fuel type.

---

### POST `/vehicles/:vehicleId/fuel` — Log a fill-up or charge

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/fuel \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "fuel_type": "petrol",
    "quantity": 40.52,
    "price_per_unit": 103.75,
    "total_cost": 4204.00,
    "odometer_reading": 45230,
    "fuel_station": "IOCL HP Pump, Kakkanad",
    "date": "2026-10-24",
    "notes": "Full tank"
  }'
```

For electric vehicles:
```json
{
  "fuel_type": "electric",
  "quantity": 45.0,
  "price_per_unit": 8.50,
  "total_cost": 382.50,
  "odometer_reading": 15800,
  "fuel_station": "Tata Power EV, InfoPark",
  "date": "2026-10-24"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `fuel_type` | string | ✅ | `petrol` `diesel` `cng` `electric` `hybrid` |
| `quantity` | float | ✅ | Litres or kWh |
| `price_per_unit` | float | ❌ | Per litre or per kWh |
| `total_cost` | float | ✅ | Total amount paid |
| `odometer_reading` | int | ❌ | km reading at fill-up |
| `fuel_station` | string | ❌ | Station name/location |
| `date` | date | ✅ | `"YYYY-MM-DD"` |
| `notes` | string | ❌ | |

### PUT `/vehicles/:vehicleId/fuel/:id` — Update log (owner/editor)
### DELETE `/vehicles/:vehicleId/fuel/:id` — Delete log (owner only)

---

## 13. Expenses

### POST `/vehicles/:vehicleId/expenses`

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/expenses \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "Accessory",
    "description": "Dashcam installation — Viofo A129 Pro",
    "amount": 8200.00,
    "date": "2026-09-20",
    "vendor": "AutoZone Kochi",
    "notes": "Includes front and rear camera"
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `category` | string | ✅ | e.g. `"Repair"` `"Accessory"` `"Cleaning"` `"Toll"` `"Parking"` `"Other"` |
| `description` | string | ✅ | |
| `amount` | float | ✅ | |
| `date` | date | ✅ | |
| `vendor` | string | ❌ | |
| `notes` | string | ❌ | |

### GET / PUT / DELETE follow the same pattern.

---

## 14. Documents

### POST `/vehicles/:vehicleId/documents` — Upload a document
`multipart/form-data`

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/documents \
  -H "Authorization: Bearer {{token}}" \
  -F "file=@/path/to/rc_book.pdf" \
  -F "document_type=RC Book" \
  -F "expiry_date=2031-06-01" \
  -F "notes=Original RC scan"
```

| Form field | Notes |
|---|---|
| `file` | PDF, JPG, PNG |
| `document_type` | e.g. `"RC Book"` `"Insurance"` `"PUC"` `"Invoice"` |
| `expiry_date` | Optional |
| `notes` | Optional |

### GET `/vehicles/:vehicleId/documents` — List documents
### DELETE `/vehicles/:vehicleId/documents/:id` — Delete (owner only)

---

## 15. Media (Photos & Videos)

### POST `/media/upload` — Attach media to any entity
`multipart/form-data`. Allowed: jpeg, png, gif, webp, mp4, mov, avi.

```bash
curl -X POST {{base}}/media/upload \
  -H "Authorization: Bearer {{token}}" \
  -F "file=@/path/to/car_photo.jpg" \
  -F "entity_type=vehicle" \
  -F "entity_id=VEHICLE_ID"
```

| Form field | Notes |
|---|---|
| `file` | Image or video |
| `entity_type` | `vehicle` `maintenance` `service` `insurance` `document` |
| `entity_id` | UUID of the related record |

### GET `/media/:entityType/:entityId` — List media for a record

```bash
curl {{base}}/media/vehicle/VEHICLE_ID -H "Authorization: Bearer {{token}}"
```

### DELETE `/media/:id`

```bash
curl -X DELETE {{base}}/media/MEDIA_ID -H "Authorization: Bearer {{token}}"
```

---

## 16. Reminders

### POST `/vehicles/:vehicleId/reminders`

```bash
curl -X POST {{base}}/vehicles/VEHICLE_ID/reminders \
  -H "Authorization: Bearer {{token}}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Insurance Renewal",
    "description": "HDFC ERGO policy renewal due",
    "due_date": "2026-12-15",
    "reminder_type": "insurance",
    "notify_days_before": 14,
    "notify_email": true,
    "notify_sms": false
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `title` | string | ✅ | |
| `description` | string | ❌ | |
| `due_date` | date | ✅ | |
| `reminder_type` | string | ❌ | `insurance` `puc` `service` `maintenance` `other` |
| `notify_days_before` | int | ❌ | Days before due to alert |
| `notify_email` | bool | ❌ | |
| `notify_sms` | bool | ❌ | |

### GET / PUT / DELETE follow the same pattern.

---

## 17. Notifications

### GET `/notifications` — My notifications (all vehicles)

```bash
curl {{base}}/notifications -H "Authorization: Bearer {{token}}"
```

---

## 18. Reports & Dashboard

### GET `/dashboard` — Summary dashboard

```bash
curl {{base}}/dashboard -H "Authorization: Bearer {{token}}"
```

Returns vehicle counts, upcoming reminders, expiring insurance/PUC, and recent activity.

---

### GET `/vehicles/:vehicleId/reports/summary` — Vehicle cost summary

```bash
curl {{base}}/vehicles/VEHICLE_ID/reports/summary \
  -H "Authorization: Bearer {{token}}"
```

---

### GET `/vehicles/:vehicleId/reports/expenses` — Expense breakdown

```bash
curl "{{base}}/vehicles/VEHICLE_ID/reports/expenses?from=2026-01-01&to=2026-12-31" \
  -H "Authorization: Bearer {{token}}"
```

---

### GET `/vehicles/:vehicleId/activity` — Activity log

```bash
curl {{base}}/vehicles/VEHICLE_ID/activity -H "Authorization: Bearer {{token}}"
```

---

## 19. Admin (Admin token required)

All `/admin/*` endpoints require `is_admin = true` on the authenticated user.

### GET `/admin/analytics` — App usage analytics

```bash
curl {{base}}/admin/analytics -H "Authorization: Bearer {{ADMIN_TOKEN}}"
```

**Response:**
```json
{
  "users": {
    "total": 142,
    "active": 138,
    "admins": 3,
    "with_email": 130,
    "with_phone": 42,
    "new_this_month": 8
  },
  "vehicles": {
    "total": 287,
    "active": 280,
    "by_type": [{ "type_name": "Sedan", "count": 120 }],
    "by_fuel_type": [{ "fuel_type": "petrol", "count": 168 }],
    "popular_brands": [{ "brand": "Toyota", "count": 45 }],
    "popular_models": [{ "brand": "Toyota", "model": "Innova", "count": 28 }]
  },
  "activity": {
    "total_maintenance_records": 1240,
    "total_service_records": 890,
    "total_insurance_records": 320,
    "total_fuel_logs": 3450,
    "total_documents": 560,
    "total_reminders": 780,
    "active_reminders": 420,
    "total_media": 1100
  }
}
```

---

### GET `/admin/admins` — List all admin users

```bash
curl {{base}}/admin/admins -H "Authorization: Bearer {{ADMIN_TOKEN}}"
```

---

### GET `/admin/users` — List all users

```bash
curl "{{base}}/admin/users?page=1&limit=20" \
  -H "Authorization: Bearer {{ADMIN_TOKEN}}"
```

---

### GET `/admin/users/:userId` — Get a specific user

```bash
curl {{base}}/admin/users/USER_ID -H "Authorization: Bearer {{ADMIN_TOKEN}}"
```

---

### PUT `/admin/users/:userId/admin` — Grant or revoke admin

```bash
# Grant admin
curl -X PUT {{base}}/admin/users/USER_ID/admin \
  -H "Authorization: Bearer {{ADMIN_TOKEN}}" \
  -H "Content-Type: application/json" \
  -d '{ "is_admin": true }'

# Revoke admin
curl -X PUT {{base}}/admin/users/USER_ID/admin \
  -H "Authorization: Bearer {{ADMIN_TOKEN}}" \
  -H "Content-Type: application/json" \
  -d '{ "is_admin": false }'
```

> A user cannot change their own admin status.

---

### PUT `/admin/users/:userId/deactivate` — Deactivate a user

```bash
curl -X PUT {{base}}/admin/users/USER_ID/deactivate \
  -H "Authorization: Bearer {{ADMIN_TOKEN}}"
```

---

### Register a new admin user (admin-only action)

```bash
curl -X POST {{base}}/auth/register \
  -H "Authorization: Bearer {{ADMIN_TOKEN}}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New Admin",
    "email": "newadmin@email.com",
    "password": "Admin@SecurePass1",
    "is_admin": true
  }'
```

> `is_admin: true` is only accepted when a valid admin `Authorization` header is provided. Without it the request returns `403`.

---

## 20. Health Check

```bash
curl http://localhost:8080/health
```

```json
{ "status": "ok" }
```

---

## Quick Reference — All Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/auth/register` | None | Create account |
| POST | `/auth/login` | None | Sign in |
| POST | `/auth/logout` | None | Sign out |
| POST | `/auth/refresh` | None | Refresh tokens |
| POST | `/auth/forgot-password` | None | Send reset email |
| POST | `/auth/reset-password` | None | Reset password with token |
| GET | `/users/me` | ✅ | Get my profile |
| PUT | `/users/me` | ✅ | Update profile |
| PUT | `/users/me/password` | ✅ | Change password |
| POST | `/users/me/avatar` | ✅ | Upload avatar |
| GET | `/vehicle-types` | ✅ | List vehicle types |
| GET | `/maintenance-types` | ✅ | List maintenance types |
| GET | `/vehicles` | ✅ | List my vehicles |
| GET | `/vehicles/search` | ✅ | Search by plate/email/phone |
| POST | `/vehicles` | ✅ | Add vehicle |
| GET | `/vehicles/:id` | ✅ | Get vehicle |
| PUT | `/vehicles/:id` | owner/editor | Update vehicle |
| DELETE | `/vehicles/:id` | owner | Archive vehicle |
| GET | `/vehicles/:id/members` | ✅ | List members |
| POST | `/vehicles/:id/invitations` | owner | Invite member |
| GET | `/invitations` | ✅ | My pending invitations |
| POST | `/invitations/:token/accept` | ✅ | Accept invitation |
| POST | `/invitations/:token/reject` | ✅ | Reject invitation |
| PUT | `/vehicles/:id/members/:uid/role` | owner | Change member role |
| DELETE | `/vehicles/:id/members/:uid` | owner | Remove member |
| GET/POST | `/vehicles/:id/maintenance` | ✅ / owner+editor | Maintenance records |
| GET/PUT/DELETE | `/vehicles/:id/maintenance/:id` | ✅ / owner+editor / owner | Single record |
| GET/POST | `/vehicles/:id/services` | ✅ / owner+editor | Service records |
| GET/PUT/DELETE | `/vehicles/:id/services/:id` | ✅ / owner+editor / owner | Single record |
| GET/POST | `/vehicles/:id/insurance` | ✅ / owner+editor | Insurance records |
| GET/PUT/DELETE | `/vehicles/:id/insurance/:id` | ✅ / owner+editor / owner | Single record |
| GET/POST | `/vehicles/:id/puc` | ✅ / owner+editor | PUC certificates |
| GET/PUT/DELETE | `/vehicles/:id/puc/:id` | ✅ / owner+editor / owner | Single record |
| GET/POST | `/vehicles/:id/tyres` | ✅ / owner+editor | Tyre records |
| GET/PUT/DELETE | `/vehicles/:id/tyres/:id` | ✅ / owner+editor / owner | Single record |
| GET/POST | `/vehicles/:id/batteries` | ✅ / owner+editor | Battery records |
| GET/PUT/DELETE | `/vehicles/:id/batteries/:id` | ✅ / owner+editor / owner | Single record |
| GET/POST | `/vehicles/:id/odometer` | ✅ / owner+editor | Odometer readings |
| GET | `/vehicles/:id/fuel/summary` | ✅ | Fuel summary |
| GET/POST | `/vehicles/:id/fuel` | ✅ / owner+editor | Fuel/charge logs |
| GET/PUT/DELETE | `/vehicles/:id/fuel/:id` | ✅ / owner+editor / owner | Single log |
| GET/POST | `/vehicles/:id/expenses` | ✅ / owner+editor | Expenses |
| GET/PUT/DELETE | `/vehicles/:id/expenses/:id` | ✅ / owner+editor / owner | Single expense |
| GET/POST | `/vehicles/:id/documents` | ✅ / owner+editor | Documents |
| GET/DELETE | `/vehicles/:id/documents/:id` | ✅ / owner | Single document |
| POST | `/media/upload` | ✅ | Upload media |
| GET | `/media/:entityType/:entityId` | ✅ | List media |
| DELETE | `/media/:id` | ✅ | Delete media |
| GET/POST | `/vehicles/:id/reminders` | ✅ / owner+editor | Reminders |
| GET/PUT/DELETE | `/vehicles/:id/reminders/:id` | ✅ / owner+editor / owner | Single reminder |
| GET | `/notifications` | ✅ | My notifications |
| GET | `/dashboard` | ✅ | Dashboard summary |
| GET | `/vehicles/:id/reports/summary` | ✅ | Vehicle cost summary |
| GET | `/vehicles/:id/reports/expenses` | ✅ | Expense report |
| GET | `/vehicles/:id/activity` | ✅ | Activity log |
| GET | `/admin/analytics` | admin | App usage analytics |
| GET | `/admin/admins` | admin | List all admins |
| GET | `/admin/users` | admin | List all users |
| GET | `/admin/users/:id` | admin | Get user |
| PUT | `/admin/users/:id/admin` | admin | Grant/revoke admin |
| PUT | `/admin/users/:id/deactivate` | admin | Deactivate user |
| GET | `/health` | None | Health check |
