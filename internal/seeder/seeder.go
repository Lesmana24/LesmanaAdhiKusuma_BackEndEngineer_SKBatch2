package seeder

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/model"
	"gorm.io/gorm"
)

// SeedApplicants inserts 10 realistic mock applicants into the database if empty
func SeedApplicants(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Applicant{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Printf("[Seeder] Database already contains %d applicant records. Skipping seeder.\n", count)
		return nil
	}

	now := time.Now()

	applicants := []model.Applicant{
		{
			ID:              uuid.NewString(),
			Name:            "Budi Santoso",
			Email:           "budi.santoso@example.com",
			Phone:           "081234567891",
			AppliedPosition: "Back-End Engineer",
			Stage:           model.StageApplied,
			Notes:           "Kandidat baru mendaftar dari job board Sekolah Karir.",
			CreatedAt:       now.Add(-72 * time.Hour),
			UpdatedAt:       now.Add(-72 * time.Hour),
		},
		{
			ID:              uuid.NewString(),
			Name:            "Siti Rahmawati",
			Email:           "siti.rahmawati@example.com",
			Phone:           "081234567892",
			AppliedPosition: "Back-End Engineer",
			Stage:           model.StageScreening,
			Notes:           "CV dan portfolio GitHub sedang direview oleh hiring manager.",
			CreatedAt:       now.Add(-48 * time.Hour),
			UpdatedAt:       now.Add(-36 * time.Hour),
		},
		{
			ID:              uuid.NewString(),
			Name:            "Ahmad Fauzi",
			Email:           "ahmad.fauzi@example.com",
			Phone:           "081234567893",
			AppliedPosition: "Front-End Engineer",
			Stage:           model.StageInterview,
			Notes:           "Jadwal wawancara user hari Jumat pukul 14:00 WIB.",
			CreatedAt:       now.Add(-96 * time.Hour),
			UpdatedAt:       now.Add(-24 * time.Hour),
		},
		{
			ID:              uuid.NewString(),
			Name:            "Jessica Tan",
			Email:           "jessica.tan@example.com",
			Phone:           "081234567894",
			AppliedPosition: "Mobile Developer",
			Stage:           model.StageTechnicalTest,
			Notes:           "Take-home test Flutter dikirimkan, batas pengumpulan 3 hari.",
			CreatedAt:       now.Add(-120 * time.Hour),
			UpdatedAt:       now.Add(-40 * time.Hour),
		},
		{
			ID:              uuid.NewString(),
			Name:            "Rizky Pratama",
			Email:           "rizky.pratama@example.com",
			Phone:           "081234567895",
			AppliedPosition: "DevOps Engineer",
			Stage:           model.StageOffered,
			Notes:           "Offering letter telah dikirimkan via email, menunggu konfirmasi.",
			CreatedAt:       now.Add(-160 * time.Hour),
			UpdatedAt:       now.Add(-12 * time.Hour),
		},
		{
			ID:              uuid.NewString(),
			Name:            "Maya Safitri",
			Email:           "maya.safitri@example.com",
			Phone:           "081234567896",
			AppliedPosition: "UI/UX Designer",
			Stage:           model.StageAccepted,
			Notes:           "Kandidat menerima tawaran magang. Onboarding tanggal 1 bulan depan.",
			CreatedAt:       now.Add(-200 * time.Hour),
			UpdatedAt:       now.Add(-10 * time.Hour),
		},
		{
			ID:              uuid.NewString(),
			Name:            "Daniel Wibowo",
			Email:           "daniel.wibowo@example.com",
			Phone:           "081234567897",
			AppliedPosition: "QA Engineer",
			Stage:           model.StageRejected,
			Notes:           "Belum memenuhi kualifikasi otomatisasi pengujian.",
			CreatedAt:       now.Add(-180 * time.Hour),
			UpdatedAt:       now.Add(-30 * time.Hour),
		},
		{
			ID:              uuid.NewString(),
			Name:            "Nadia Putri",
			Email:           "nadia.putri@example.com",
			Phone:           "081234567898",
			AppliedPosition: "Back-End Engineer",
			Stage:           model.StageInterview,
			Notes:           "Lolos tes algoritma, lanjut wawancara kultur tim engineering.",
			CreatedAt:       now.Add(-60 * time.Hour),
			UpdatedAt:       now.Add(-15 * time.Hour),
		},
		{
			ID:              uuid.NewString(),
			Name:            "Hendra Wijaya",
			Email:           "hendra.wijaya@example.com",
			Phone:           "081234567899",
			AppliedPosition: "Data Analyst",
			Stage:           model.StageScreening,
			Notes:           "Pengalaman analisa data SQL dan dashboard Looker baik.",
			CreatedAt:       now.Add(-30 * time.Hour),
			UpdatedAt:       now.Add(-20 * time.Hour),
		},
		{
			ID:              uuid.NewString(),
			Name:            "Kevin Sanjaya",
			Email:           "kevin.sanjaya@example.com",
			Phone:           "081234567800",
			AppliedPosition: "Back-End Engineer",
			Stage:           model.StageApplied,
			Notes:           "Kandidat mahasiswa tingkat akhir dengan proyek Golang Gin.",
			CreatedAt:       now.Add(-5 * time.Hour),
			UpdatedAt:       now.Add(-5 * time.Hour),
		},
	}

	if err := db.CreateInBatches(applicants, 10).Error; err != nil {
		return err
	}

	log.Printf("[Seeder] Successfully seeded %d mock applicant records into database.\n", len(applicants))
	return nil
}

