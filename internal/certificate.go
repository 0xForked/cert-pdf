package internal

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/signintech/gopdf"
)

type PreGenerateCertificate struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	SchoolID        string          `json:"school_id"`
	MaterialID      string          `json:"material_id"`
	Collection      string          `json:"type"`
	CollectionID    string          `json:"type_id,omitempty"`
	ReferenceNumber string          `json:"reference_number"`
	Metadata        *CourseMetadata `json:"metadata"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       *time.Time      `json:"updated_at"`
	ValidFrom       *time.Time      `json:"valid_from"`
	ValidUntil      *time.Time      `json:"valid_until"`
	// relation data
	Material *Material          `json:"material"`
	Talent   *User              `json:"talent"`
	Course   *CourseProgramItem `json:"course,omitempty"`
	Program  *CourseProgramItem `json:"program,omitempty"`
}

type Material struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	Signature *string `json:"admin_signature"`
	Owner     *Owner  `json:"owner_certificate"`
	School    *School `json:"school"`
}

type Owner struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Email       string `json:"email,omitempty"`
	JobTitle    string `json:"job_title,omitempty"`
}

type School struct {
	ID          string     `json:"_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Location    string     `json:"location"`
	Logo        string     `json:"logo"`
	SocialURL   *SocialURL `json:"social_url,omitempty"`
}

type SocialURL struct {
	Website   string `json:"website_url"`
	LinkedIn  string `json:"linkedin_url"`
	Twitter   string `json:"twitter_url"`
	Facebook  string `json:"facebook_url"`
	Instagram string `json:"instagram_url"`
}

type CourseProgramItem struct {
	ID   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}

type CourseMetadata struct {
	CourseID     string                 `json:"course_id"`
	GreenSkillID string                 `json:"green_skill_id"`
	Name         string                 `json:"name"`
	Levels       []*CourseLevelMetadata `json:"levels"`
}

type CourseLevelMetadata struct {
	ID              string     `json:"id"`
	ImageURL        string     `json:"image_url"`
	LearningOutcome string     `json:"learning_outcome"`
	Name            string     `json:"name"`
	Proficiency     int        `json:"proficiency"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	Avatar    string `json:"avatar"`
	Location  string `json:"location"`
}

func (c PreGenerateCertificate) GeneratePDF(path string) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4Landscape})
	pdf.AddPage()

	if err := pdf.AddTTFFont("kodchasan_regular", "./assets/fonts/kodchasan/Regular.ttf"); err != nil {
		return err
	}
	if err := pdf.AddTTFFont("sen_regular", "./assets/fonts/sen/Regular.ttf"); err != nil {
		return err
	}
	if err := pdf.AddTTFFont("rammettoone_regular", "./assets/fonts/rammettoone/Regular.ttf"); err != nil {
		return err
	}
	if err := pdf.AddTTFFont("sofia_regular", "./assets/fonts/sofia/Regular.ttf"); err != nil {
		return err
	}

	// school image - 200 x 130 - gap to bottom 16px
	filePath, err := downloadImageFile(
		filepath.Join(".", "assets", "images"),
		c.Material.School.Logo)
	if err != nil {
		return err
	}
	imgFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening image: %v", err)
	}
	defer func() { _ = imgFile.Close() }()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return fmt.Errorf("error decoding image: %v", err)
	}
	imgWidth := float64(img.Bounds().Dx())
	imgHeight := float64(img.Bounds().Dy())
	desiredWidth := 200.0
	desiredHeight := 130.0
	aspectRatio := imgWidth / imgHeight
	if desiredWidth/aspectRatio > desiredHeight {
		desiredWidth = desiredHeight * aspectRatio
	} else {
		desiredHeight = desiredWidth / aspectRatio
	}
	xPos := (gopdf.PageSizeA4Landscape.W - desiredWidth) / 2
	if err := pdf.Image(filePath, xPos, 30, &gopdf.Rect{W: desiredWidth, H: desiredHeight}); err != nil {
		return fmt.Errorf("error adding image to pdf: %v", err)
	}

	startY := 30 + desiredHeight + 20

	setFont := func(style string, fontSize int) error {
		switch style {
		case "sen_regular":
			return pdf.SetFont("sen_regular", "", fontSize)
		case "rammettoone_regular":
			return pdf.SetFont("rammettoone_regular", "", fontSize)
		case "sofia_regular":
			return pdf.SetFont("sofia_regular", "", fontSize)
		default:
			return pdf.SetFont("kodchasan_regular", "", fontSize)
		}
	}

	centerText := func(text, style string, y float64, fontSize int, r, g, b uint8) (float64, float64, error) {
		if err := setFont(style, fontSize); err != nil {
			return 0, 0, err
		}
		pdf.SetTextColor(r, g, b)
		textWidth, err := pdf.MeasureTextWidth(text)
		if err != nil {
			return 0, 0, err
		}
		x := (gopdf.PageSizeA4Landscape.W - textWidth) / 2
		pdf.SetXY(x, y)
		if err := pdf.Text(text); err != nil {
			return 0, 0, err
		}
		pdf.SetTextColor(0, 0, 0)
		return x, textWidth, nil
	}

	if _, _, err := centerText(
		"C E R T I F I C A T E   O F   C O M P L E T I O N",
		"sen_regular", startY, 12, 150, 111, 2,
	); err != nil {
		return err
	}

	if _, _, err := centerText(
		"This is to certify that", "regular",
		startY+20, 10, 0, 0, 0,
	); err != nil {
		return err
	}

	yPos := startY + 50
	fontSize := 16
	startX, textWidth, err := centerText(c.Talent.FullName,
		"sofia_regular", yPos, 16, 0, 0, 0) // 120
	if err != nil {
		return err
	}
	underlineY := yPos + float64(fontSize) - 10
	pdf.SetLineWidth(0.5)
	pdf.Line(startX, underlineY, startX+textWidth, underlineY)

	if _, _, err := centerText(
		"has successfully completed online [non-credit]",
		"regular", startY+80, 10, 0, 0, 0, // 150
	); err != nil {
		return err
	}

	if _, _, err := centerText(
		strings.ToUpper(c.Metadata.Name), "rammettoone_regular",
		startY+115, 18, 0, 0, 0, // 185
	); err != nil {
		return err
	}

	if _, _, err := centerText(
		"A course delivered by School Name and offered through Skilledin Green platform",
		"regular", startY+145, 10, 0, 0, 0, // 215
	); err != nil {
		return err
	}

	pdf.SetXY(30, startY+180) //250
	if err := pdf.Text("Completed proficiency levels and associated knowledge:"); err != nil {
		return err
	}

	nextY := startY + 185.0
	const (
		col1Width = 40
		col2Width = 60
		col3Width = 100
		maxWords  = 10
	)
	for _, level := range c.Metadata.Levels {
		nextY += 25
		pdf.SetXY(30, nextY)
		updatedAt := level.UpdatedAt.Format("02 January 2006")
		proficiency := fmt.Sprintf("Level %d (%s)", level.Proficiency, capitalizeFirst(level.Name))
		learningOutcome := level.LearningOutcome

		if err := pdf.Text(updatedAt); err != nil {
			return err
		}
		pdf.SetXY(pdf.GetX()+col1Width, nextY)

		if err := pdf.Text(proficiency); err != nil {
			return err
		}
		pdf.SetXY(pdf.GetX()+col2Width, nextY)

		// Handle the learning outcome with splitting into lines of maxWords
		learningOutcomeY := nextY
		words := strings.Fields(learningOutcome) // Split the string into words
		for i := 0; i < len(words); i += maxWords {
			// Get the next part of the learning outcome
			end := i + maxWords
			if end > len(words) {
				end = len(words) // Ensure we don't go out of bounds
			}
			part := strings.Join(words[i:end], " ") // Join the words back into a string
			// Print the learning outcome part
			if err := pdf.Text(part); err != nil {
				return err
			}
			// Move down for the next part
			learningOutcomeY += 10
			pdf.SetXY((col2Width+col3Width)*2, learningOutcomeY)
		}
	}

	return pdf.WritePdf(fmt.Sprintf("%s/%s.pdf", path, c.ReferenceNumber))
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func downloadImageFile(dirPath string, url string) (string, error) {
	// Extract the file name from the URL
	fileName := path.Base(url)
	// Create the full file path
	fullFilePath := filepath.Join(dirPath, fileName)
	// Download the image
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	// Create the file
	out, err := os.Create(fullFilePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = out.Close() }()
	// Write the body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	return fullFilePath, nil
}
