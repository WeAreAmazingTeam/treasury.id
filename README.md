# Technical Test (treasury.id)

Jawaban untuk lima soal di [TECHNICAL_TEST.txt](TECHNICAL_TEST.txt). Project ini menggunakan Go 1.27.1 dengan nama module `treasury.id`, sesuai [go.mod](go.mod).

[main.go](main.go) menjalankan semua contoh soal menggunakan versi `simple`.

## Cara menjalankan

Jalankan dari folder utama project. Pastikan Go 1.27.1 sudah terpasang.

```bash
go version
go mod download
go run .
```

Hasilnya seperti ini.

```text
Soal 1:
0
-2
3

Soal 2:
0 not found
62952 <nil>

Soal 3:
[1]
[2 1]

Soal 4:
1 -> 1
2 -> 2
3 -> 3
4 -> 1

Soal 5:
14
39
```

## Versi yang tersedia

Ada tiga versi dengan nama fungsi, parameter, dan hasil yang sama. Setiap folder berisi `solution.go` dan `solution_test.go`.

- [optimized](optimized/solution.go) adalah versi awal yang mengutamakan waktu proses dan penggunaan memori. Pencarian angka memakai cyclic placement, yaitu menempatkan angka ke posisi yang sesuai dalam slice. Pengolahan digit memakai operasi matematika. Pengurutan memakai standard library untuk hasil kecil dan radix sort untuk hasil besar. Perkalian pasangan memakai goroutine dan `sync.WaitGroup`. Test ada di [optimized/solution_test.go](optimized/solution_test.go).
- [simple](simple/solution.go) memakai map, `strings.Builder`, konversi angka, dan sorting dari standard library. Hasil perkalian dikirim melalui channel dengan buffer. Ini versi yang dipakai oleh `main.go`. Test ada di [simple/solution_test.go](simple/solution_test.go).
- [concise](concise/solution.go) memakai helper dari `samber/lo`, seperti `lo.Map`, `lo.Filter`, `lo.Intersect`, dan `lo.CountValues`. Pembalikan slice memakai `mutable.Reverse`. Perkalian pasangan memakai `lo.Async`, lalu hasilnya dijumlahkan dengan `lo.Reduce`. Test ada di [concise/solution_test.go](concise/solution_test.go).

Versi `optimized` dan `simple` memakai standard library Go. Versi `concise` memakai [samber/lo v1.53.0](https://github.com/samber/lo/tree/v1.53.0). Versi library sudah ditentukan di `go.mod`. Library ini juga membawa `golang.org/x/text v0.22.0` sebagai dependency tidak langsung. Checksum keduanya ada di [go.sum](go.sum).

## Penjelasan soal

1. **Mencari angka yang belum ada**

   Fungsi `Closest(arr []int) int` mencari angka yang belum ada dan paling dekat ke nol, mengikuti contoh soal. Untuk input positif, pencarian dimulai dari `1`. Untuk input negatif, pencarian dimulai dari `-1`. Kalau ada angka positif dan negatif sekaligus, hasilnya `0`. Nol tidak ikut menentukan tanda input.

   Contohnya, `[-1,1]` menghasilkan `0`, `[-1,-7,-5]` menghasilkan `-2`, dan `[1,2,1,6]` menghasilkan `3`.

   Pada versi `simple`, semua angka dicatat ke map. Setelah itu, fungsi memeriksa calon angka satu per satu sampai menemukan angka yang belum tercatat.

2. **Menggabungkan dan membalik angka**

   Fungsi `JoinExcluding(exclude []int, nums ...int) (int, error)` menggabungkan angka menjadi satu integer. Digitnya lalu dibaca dari belakang, sambil melewati digit yang ada dalam daftar exclude.

   Contohnya, `1259` dan `2601` digabung menjadi `12592601`. Setelah dibalik menjadi `10629521`, digit `0` dan `1` dibuang. Hasil akhirnya `62952` dengan error `nil`.

   Untuk `1000` dan `1010` dengan exclude yang sama, tidak ada digit yang tersisa. Hasilnya `0` dengan error `"not found"`.

   Versi `simple` memakai `strings.Builder` untuk menyusun teks angka dan `strconv` untuk mengubah teks menjadi integer. Daftar exclude disimpan dalam array `[10]bool`, satu posisi untuk setiap digit dari `0` sampai `9`.

   Gabungan diubah menjadi integer sebelum dibalik. Jadi input `0` dan `12` menjadi `12`, lalu hasil baliknya `21`.

3. **Mencari angka yang sama pada dua daftar**

   Fungsi `UniqueDescending(numsA, numsB []int) []int` mengambil angka yang ada di kedua daftar. Setiap angka hanya dimasukkan sekali, lalu hasilnya diurutkan dari terbesar ke terkecil.

   Daftar `[1,2]` dan `[1,3]` menghasilkan `[1]`. Daftar `[1,2,2]` dan `[1,2,4]` menghasilkan `[2,1]`.

   Versi `simple` mencatat isi daftar yang lebih pendek ke map, lalu memeriksa daftar satunya. Angka yang sudah masuk hasil dihapus dari map agar tidak terambil dua kali. Hasilnya diurutkan dengan `sort.Reverse` dan `sort.Sort`.

4. **Menghitung kemunculan digit**

   Fungsi `Count(num int) map[int]int` membagi angka menjadi array atau slice digit integer, lalu mengelompokkan dan menghitung digit yang sama. Angka `1223334` menjadi `[1,2,2,3,3,3,4]`. Dari daftar itu, digit `1` dihitung sekali, `2` dua kali, `3` tiga kali, dan `4` sekali.

   Versi `simple` mengambil digit terakhir dengan `% 10` dan menyimpannya ke slice. Pembagian integer `/ 10` membuang digit yang sudah diambil. Setelah semua digit terkumpul, slice dibalik agar urutannya sama dengan angka awal. Barulah setiap digit dihitung ke map.

   Semua versi mengikuti dua tahap tersebut. Versi `optimized` memakai array tetap dengan kapasitas 19 digit, cukup untuk nilai `int` 32 atau 64 bit. Versi `concise` membentuk slice integer dengan `lo.Map`, lalu menghitungnya dengan `lo.CountValues`.

5. **Menjumlahkan hasil perkalian pasangan**

   Fungsi `SumProducts(nums ...int) int` mengalikan angka pada indeks `0` dan `1`, lalu `2` dan `3`, dan seterusnya. Angka terakhir yang tidak punya pasangan dikalikan dengan dirinya sendiri.

   Input `1,2,3,4` menghasilkan `14` dari `(1*2) + (3*4)`. Ada dua goroutine untuk menghitung perkaliannya. Input `1,2,3,4,5` menghasilkan `39` karena ditambah `5*5`, dengan tiga goroutine perkalian.

   Versi `simple` memakai satu goroutine untuk setiap pasangan. Hasil perkalian dikirim lewat channel. Goroutine pemanggil menerima semua hasil dan menjumlahkannya. Jadi penjumlahan tetap dilakukan oleh pemanggil, sesuai soal.

## Big O

Versi `optimized`.

| Soal | Waktu | Memori |
| --- | --- | --- |
| 1 | `O(n)` | `O(1)` |
| 2 | `O(e+d)` | `O(1)` |
| 3 | Rata-rata `O(a+b+k)` untuk integer 32 atau 64 bit | `O(u+k)` |
| 4 | `O(d)` | `O(1)` |
| 5 | `O(n)` untuk seluruh pekerjaan | `O(p)` |

Versi `simple`.

| Soal | Waktu | Memori |
| --- | --- | --- |
| 1 | Rata-rata `O(n)` | `O(u)` |
| 2 | `O(e+d)` | `O(d)` |
| 3 | Rata-rata `O(a+b+k log k)` | `O(u+k)` |
| 4 | `O(d)` | `O(d)` |
| 5 | `O(n)` untuk seluruh pekerjaan | `O(p)` |

`n` adalah panjang input, `e` adalah jumlah elemen exclude, dan `d` adalah jumlah digit yang diproses. Untuk soal 3, `a` dan `b` adalah panjang kedua daftar, `k` adalah jumlah hasil unik, dan `u` adalah jumlah angka unik yang disimpan dalam map. Map pada soal 3 berasal dari daftar yang lebih pendek. `p` adalah jumlah pasangan, termasuk angka terakhir yang tidak punya pasangan.

Perhitungan memori mencakup hasil yang dikembalikan dan goroutine yang dibuat. Waktu pencarian dengan map memakai perkiraan rata-rata. Map penghitung digit punya paling banyak 10 entri, sehingga memorinya tetap `O(1)`. Slice digit pada versi `simple` membutuhkan `O(d)` memori. Versi `optimized` memakai array dengan kapasitas tetap, sehingga memorinya tetap `O(1)`.

Pada `optimized`, hasil irisan yang sudah terurut langsung dipakai atau dibalik. Hasil yang belum terurut dengan paling banyak 1024 angka memakai sorting standard library. Hasil yang lebih besar memakai radix sort berdasarkan byte, dengan 4 putaran untuk integer 32 bit atau 8 putaran untuk integer 64 bit.

## Catatan penggunaan

- `Closest` menghasilkan `0` untuk input kosong atau input yang hanya berisi nol. Angka yang berulang tidak mengubah hasil pencarian.
- `Closest` pada versi `optimized` bisa mengubah urutan input. Gunakan `slices.Clone(arr)` kalau urutan awal perlu disimpan. Salinan ini membutuhkan memori tambahan sesuai panjang input. Fungsi lain pada `optimized`, serta seluruh fungsi pada `simple` dan `concise`, tidak mengubah input.
- `JoinExcluding` menghasilkan `ErrNotFound` kalau tidak ada angka yang diberikan atau semua digit terbuang. Kalau masih ada digit nol, hasilnya `0` dengan error `nil`.
- Angka negatif pada `JoinExcluding` menghasilkan `ErrNegativeNumber`. Gabungan atau hasil yang melebihi kapasitas `int` menghasilkan `ErrOverflow`. Batas gabungan diperiksa sebelum digit dibuang.
- Digit exclude yang berulang tidak memengaruhi hasil. Nilai exclude di luar `0` sampai `9` diabaikan.
- `UniqueDescending` menghasilkan `nil` kalau kedua daftar tidak punya angka yang sama.
- `Count` mengabaikan tanda minus dan mendukung `MinInt`. Input `0` menghasilkan `map[int]int{0: 1}`. Urutan iterasi map bisa berbeda, jadi contoh di `main.go` mencetak digit satu per satu.
- `SumProducts` menghasilkan `0` untuk input kosong. Perkalian dan penjumlahannya memakai aturan overflow bawaan tipe `int` Go.

Setiap package punya error sendiri dengan pesan berikut.

| Error | Pesan |
| --- | --- |
| `ErrNotFound` | `not found` |
| `ErrNegativeNumber` | `negative numbers are not supported` |
| `ErrOverflow` | `integer overflow` |

Untuk memeriksa error, gunakan `errors.Is` dengan error dari package yang dipanggil. Misalnya `errors.Is(err, simple.ErrNotFound)` saat memakai versi `simple`.

## Test dan lint

Jalankan dari folder utama project.

```bash
go test ./...
go test -race -cover ./...
go vet ./...
gofmt -l .
```

Test memakai semua contoh soal, input kosong, angka berulang, nol, angka negatif, dan batas nilai integer. Ada juga test untuk input besar, dua slice yang memakai array yang sama, serta beberapa pemanggilan fungsi secara bersamaan.

Kalau hanya ingin menguji versi yang dipakai `main.go`, jalankan `go test ./simple`. Kalau formatting sudah sesuai, `gofmt -l .` tidak mencetak nama file.

Untuk lint, jalankan perintah berikut kalau `golangci-lint` sudah terpasang.

```bash
golangci-lint run ./...
```

## Fuzz test

Fuzz test mencoba banyak variasi input untuk mencari kasus yang belum tercakup oleh test biasa. `FuzzAgreement` membandingkan hasil ketiga versi. Panjang input dibatasi sampai 128 elemen.

```bash
go test ./simple -run='^$' -fuzz='^FuzzAgreement$' -fuzztime=5s -parallel=2
```

Versi `optimized` punya enam target tambahan. Jalankan satu target dalam setiap perintah.

```bash
go test ./optimized -run='^$' -fuzz='^FuzzClosest$' -fuzztime=5s -parallel=2
go test ./optimized -run='^$' -fuzz='^FuzzJoinExcluding$' -fuzztime=5s -parallel=2
go test ./optimized -run='^$' -fuzz='^FuzzSingleNumber$' -fuzztime=5s -parallel=2
go test ./optimized -run='^$' -fuzz='^FuzzUniqueDescending$' -fuzztime=5s -parallel=2
go test ./optimized -run='^$' -fuzz='^FuzzSortDescending$' -fuzztime=5s -parallel=2
go test ./optimized -run='^$' -fuzz='^FuzzCount$' -fuzztime=5s -parallel=2
```
