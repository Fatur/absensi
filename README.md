# Absensi
Mengapa disebut absensi untuk mendata kehadiran seseorang? Bukan presensi? Atau malah mungkin attandance. Saya juga kurang tahu.
#### WTF is Absensi?
Project ini ditujukan untuk mengolah data event swap seorang karyawan. Ketika seorang karyawan sampai dikantor, kemudian menggesekkan ID cardnya atau jempolnya ke sebuah mesin, saat itulah terjadi event swap. Aktifitas ini menghasilkan data swap event: EmployeeId, Location, Date, Time, Type

Type berisi tipe event, masuk atau keluar.

Data ini kemudian kita olah menjadi data kehadiran atau attandance. Data kehadiran berisi tanggal berapa dia hadir, mulai jam berapa dan sampai jam berapa.
####Akses ke system
Ada dua cara untuk mengirim perintah ke system: command lewat API dan upload file csv.

Kemudian data bisa dieksport balik dalam bentuk CSV atau JSON lewat query.
