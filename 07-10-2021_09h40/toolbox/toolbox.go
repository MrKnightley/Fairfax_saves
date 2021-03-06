package toolbox

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// BRAND NEW TEST (3) GIT IGNORE

// Fonction vΓ©rifiant si une input est vide ou non :
func IsEmptyString(str string) bool {
	for _, char := range str {
		if char != ' ' && char != '	' && char != '\t' && char != '\n' {
			return false // OKAY, la string n'est pas vide
		}
	}
	return true // NOT OKAY, la string est vide
}

// Fonction rΓ©cupΓ©rant un ππ situΓ© Γ  la fin d'un URL :
// L'URL sera de la forme π₯π¨πππ₯π‘π¨π¬π­:ππππ/πππ­ππ π¨π«π²/{ππ} ou encore π₯π¨πππ₯π‘π¨π¬π­:ππππ/π©π¨π¬π­/{ππ} selon l'URL.
func ParseURL(w http.ResponseWriter, r *http.Request) (int, error) {
	URL := r.URL.Path                                // URL complet aprΓ¨s le numΓ©ro de port (c'est-Γ -dire /πππ­ππ π¨π«π²/{ππ} ou /π©π¨π¬π­/{ππ}, etc.)
	index_2nd_Slash := strings.Index(URL[1:], "/")   // Indice du 2nd slash
	ID, err := strconv.Atoi(URL[index_2nd_Slash+2:]) // ID aprΓ¨s le 2nd slash (converti en int)
	if err != nil || ID < 1 {
		fmt.Println("β TOOLBOX | ERREUR : Impossible de rΓ©cupΓ©rer l'ID depuis r.URL.Path : ", URL)
		return ID, err
	}
	return ID, nil
}

// Fonction permettant Γ  l'utilisateur d'uploader une image sur le serveur :
func UploadImage(r *http.Request, userID int, postOrAvatar string) (string, error) {

	// (1) Lecture (parsing) du fichier :
	myFile, myFileHeader, err := r.FormFile("image")
	if err != nil {
		log.Println("β UPLOAD | Γchec de l'upload du fichier.")
		return "", err
	}

	defer myFile.Close()

	// (2) VΓ©rification de la taille du fichier (infΓ©rieure Γ  5 Mo, c-Γ -d 5*1024*1024 Ko) :
	if myFileHeader.Size > 5*1024*1024 {
		log.Println("β UPLOAD | Γchec de l'upload du fichier : fichier trop lourd (max : 5 Mo).")
		return "", errors.New("The file cannot be larger than 5 Mb.")
	}

	// (3) Lecture des 512 premiers bits du fichier pour vΓ©rifier qu'il n'est pas corrompu :
	buff := make([]byte, 512)
	_, err = myFile.Read(buff)
	if err != nil {
		log.Println("β UPLOAD | Γchec de l'upload du fichier : le fichier est corrumpu.")
		return "", errors.New("The file is corrupted.")
	}

	// (4) VΓ©rification qu'il s'agit bien d'un fichier image :
	mimeType := http.DetectContentType(buff)   // Dans le cas d'une image, le type MIME est 'image/jpeg', 'image/gif', 'image/ico', etc.
	if !strings.HasPrefix(mimeType, "image") { // Si le type MIME ne commence pas par 'image', ce n'est pas une image :
		log.Println("β UPLOAD | Γchec de l'upload du fichier : le fichier n'est pas du type βimageβ.")
		return "", errors.New("The file is not a valid image.")
	}

	// (5) VΓ©rification qu'il s'agit bien d'un fichier image π?ππ²π° ππ»π² π²πππ²π»ππΆπΌπ» ππ?πΉπΆπ±π² :
	var fileExtension string
	extensions := []string{".jpg", ".JPG", ".JPEG", ".jpeg", ".jpe", ".png", ".PNG", ".gif", ".jif", ".webp", ".ico"}

	for _, value := range extensions {
		if strings.HasSuffix(myFileHeader.Filename, value) {
			fileExtension = value
			break
		}
	}
	if fileExtension == "" {
		log.Println("β UPLOAD | Γchec de l'upload du fichier : le fichier image n'a aucune extension valide.")
		return "", errors.New("The image file has no valid extension.")
	}

	// (6) Enregistrement de l'image (au format 'π?π¬ππ«-{ππ}__{ππππ}.ππ±π­ππ§π¬π’π¨π§') :
	fileName := "user-" + strconv.Itoa(userID) + "__" + uuid.New().String()
	var imagePath string

	// 1οΈβ£ Si l'on souhaite uploader une image pour un post :
	if postOrAvatar == "post" {
		imagePath = "/images/posts/" + fileName + fileExtension // Par exemple : "/images/post/user-27__123e4567-e89b-12d3-a456-426614174000.jpg"
	}

	// 2οΈβ£ Si l'on souhaite uploader un avatar :
	if postOrAvatar == "avatar" {
		imagePath = "/images/avatars/" + fileName + fileExtension // Par exemple : "/images/post/user-27__123e4567-e89b-12d3-a456-426614174000.jpg"
	}

	newImage, err := os.Create("./database" + imagePath)
	if err != nil {
		log.Println("β UPLOAD | Γchec de l'enregistrement de l'image vers le chemin suivant : ./database", imagePath)
		return "", err
	}

	myFile.Seek(0, 0)
	io.Copy(newImage, myFile)

	return imagePath, nil
}
