package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"

	_ "github.com/go-sql-driver/mysql"
)

// Definisikan struktur data yang sesuai dengan tabel di database
type Alat struct {
	ID               int
	NamaAlat         string
	Deskripsi        string
	JumlahTotal      int
	JumlahTersedia   int
	HargaSewaPerHari float64
}

// Global variables
var tmpl *template.Template
var db *sql.DB

// Fungsi untuk koneksi ke database
func initDB() {
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASS", "")
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "db_rental_umkm")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Gagal membuka koneksi database:", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Gagal terkoneksi ke database:", err)
	}
	fmt.Println("Sukses terkoneksi ke database MySQL!")
}

// Fungsi helper untuk mendapatkan environment variable
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Fungsi untuk parsing semua file template
func initTemplates() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

// Handler untuk halaman utama, sekarang langsung ke dashboard
func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Handler untuk menampilkan dashboard
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, nama_alat, deskripsi, jumlah_total, jumlah_tersedia, harga_sewa_per_hari FROM alat ORDER BY id DESC")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var daftarAlat []Alat
	for rows.Next() {
		var alat Alat
		err := rows.Scan(&alat.ID, &alat.NamaAlat, &alat.Deskripsi, &alat.JumlahTotal, &alat.JumlahTersedia, &alat.HargaSewaPerHari)
		if err != nil {
			log.Println(err)
			continue
		}
		daftarAlat = append(daftarAlat, alat)
	}

	data := map[string]interface{}{
		"DaftarAlat": daftarAlat,
	}
	tmpl.ExecuteTemplate(w, "dashboard.html", data)
}

// Handler untuk form tambah alat
func tambahAlatHandler(w http.ResponseWriter, r *http.Request) {
	// Data dummy bisa dikosongkan karena tidak ada lagi info user
	tmpl.ExecuteTemplate(w, "tambah_alat.html", nil)
}

// Handler untuk proses tambah alat
func prosesTambahAlatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}
	namaAlat := r.FormValue("nama_alat")
	deskripsi := r.FormValue("deskripsi")
	jumlahTotal, _ := strconv.Atoi(r.FormValue("jumlah_total"))
	hargaSewa, _ := strconv.ParseFloat(r.FormValue("harga_sewa_per_hari"), 64)

	_, err := db.Exec("INSERT INTO alat (nama_alat, deskripsi, jumlah_total, jumlah_tersedia, harga_sewa_per_hari) VALUES (?, ?, ?, ?, ?)",
		namaAlat, deskripsi, jumlahTotal, jumlahTotal, hargaSewa)
	if err != nil {
		log.Println("Gagal menyimpan data:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Tidak ada lagi flash message, langsung redirect
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Handler untuk form edit alat
func editAlatHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID tidak ditemukan", http.StatusBadRequest)
		return
	}
	id, _ := strconv.Atoi(idStr)

	var alat Alat
	err := db.QueryRow("SELECT id, nama_alat, deskripsi, jumlah_total, jumlah_tersedia, harga_sewa_per_hari FROM alat WHERE id = ?", id).Scan(&alat.ID, &alat.NamaAlat, &alat.Deskripsi, &alat.JumlahTotal, &alat.JumlahTersedia, &alat.HargaSewaPerHari)
	if err != nil {
		log.Println("Data tidak ditemukan:", err)
		http.NotFound(w, r)
		return
	}
	data := map[string]interface{}{
		"Alat": alat,
	}
	tmpl.ExecuteTemplate(w, "edit_alat.html", data)
}

// Handler untuk proses update alat
func prosesUpdateAlatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	namaAlat := r.FormValue("nama_alat")
	deskripsi := r.FormValue("deskripsi")
	jumlahTotal, _ := strconv.Atoi(r.FormValue("jumlah_total"))
	jumlahTersedia, _ := strconv.Atoi(r.FormValue("jumlah_tersedia"))
	hargaSewa, _ := strconv.ParseFloat(r.FormValue("harga_sewa_per_hari"), 64)

	_, err := db.Exec("UPDATE alat SET nama_alat=?, deskripsi=?, jumlah_total=?, jumlah_tersedia=?, harga_sewa_per_hari=? WHERE id=?",
		namaAlat, deskripsi, jumlahTotal, jumlahTersedia, hargaSewa, id)
	if err != nil {
		log.Println("Gagal update data:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Handler untuk menghapus alat
func hapusAlatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	_, err := db.Exec("DELETE FROM alat WHERE id = ?", id)
	if err != nil {
		log.Println("Gagal menghapus data:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Handler untuk generate laporan
func reportHandler(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	if format == "pdf" {
		generatePdfReport(w, r)
	} else if format == "excel" {
		generateExcelReport(w, r)
	} else {
		http.Error(w, "Format tidak didukung", http.StatusBadRequest)
	}
}

// Fungsi untuk membuat laporan PDF
func generatePdfReport(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, nama_alat, deskripsi, jumlah_total, jumlah_tersedia, harga_sewa_per_hari FROM alat ORDER BY id ASC")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Laporan Daftar Alat Rental UMKM")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 10, "Tanggal Cetak: "+time.Now().Format("02 January 2006"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	headers := []string{"ID", "Nama Alat", "Deskripsi", "Stok", "Harga/Hari"}
	colWidths := []float64{15, 60, 75, 20, 25}
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 7, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(255, 255, 255)
	for rows.Next() {
		var alat Alat
		err := rows.Scan(&alat.ID, &alat.NamaAlat, &alat.Deskripsi, &alat.JumlahTotal, &alat.JumlahTersedia, &alat.HargaSewaPerHari)
		if err != nil {
			log.Println(err)
			continue
		}

		stokStr := fmt.Sprintf("%d/%d", alat.JumlahTersedia, alat.JumlahTotal)
		hargaStr := fmt.Sprintf("%.0f", alat.HargaSewaPerHari)

		pdf.CellFormat(colWidths[0], 7, strconv.Itoa(alat.ID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[1], 7, alat.NamaAlat, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[2], 7, alat.Deskripsi, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[3], 7, stokStr, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[4], 7, hargaStr, "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=laporan-alat-umkm.pdf")

	if err := pdf.Output(w); err != nil {
		log.Println("Gagal generate PDF:", err)
		http.Error(w, "Gagal membuat PDF", http.StatusInternalServerError)
	}
}

// Fungsi untuk membuat laporan Excel
func generateExcelReport(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, nama_alat, deskripsi, jumlah_total, jumlah_tersedia, harga_sewa_per_hari FROM alat ORDER BY id ASC")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	f := excelize.NewFile()
	sheetName := "Daftar Alat"
	f.SetSheetName("Sheet1", sheetName)

	// Set header tabel
	headers := []string{"ID", "Nama Alat", "Deskripsi", "Jumlah Total", "Jumlah Tersedia", "Harga Sewa/Hari"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Isi data dari database
	rowNum := 2
	for rows.Next() {
		var alat Alat
		err := rows.Scan(&alat.ID, &alat.NamaAlat, &alat.Deskripsi, &alat.JumlahTotal, &alat.JumlahTersedia, &alat.HargaSewaPerHari)
		if err != nil {
			log.Println(err)
			continue
		}
		f.SetCellValue(sheetName, "A"+strconv.Itoa(rowNum), alat.ID)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(rowNum), alat.NamaAlat)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(rowNum), alat.Deskripsi)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(rowNum), alat.JumlahTotal)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(rowNum), alat.JumlahTersedia)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(rowNum), alat.HargaSewaPerHari)
		rowNum++
	}

	// Set header untuk download file
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=laporan-alat-umkm.xlsx")

	// Tulis output Excel ke ResponseWriter
	if err := f.Write(w); err != nil {
		log.Println("Gagal generate Excel:", err)
		http.Error(w, "Gagal membuat Excel", http.StatusInternalServerError)
	}
}

func main() {
	// Inisialisasi
	initDB()
	initTemplates()
	defer db.Close()

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routing (semua rute sekarang publik, tanpa otentikasi)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/alat/tambah", tambahAlatHandler)
	http.HandleFunc("/alat/proses-tambah", prosesTambahAlatHandler)
	http.HandleFunc("/alat/edit", editAlatHandler)
	http.HandleFunc("/alat/proses-update", prosesUpdateAlatHandler)
	http.HandleFunc("/alat/hapus", hapusAlatHandler)
	http.HandleFunc("/report", reportHandler)

	// Menjalankan server
	port := getEnv("PORT", "8085")
	fmt.Printf("Server berjalan di http://localhost:%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
