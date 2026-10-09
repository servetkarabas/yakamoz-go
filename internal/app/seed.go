package app

import (
	"context"
	"log/slog"

	"github.com/karabas/yakamoz/internal/author"
	"github.com/karabas/yakamoz/internal/comment"
	"github.com/karabas/yakamoz/internal/config"
	"github.com/karabas/yakamoz/internal/platform/ai"
	"github.com/karabas/yakamoz/internal/topic"
	"github.com/karabas/yakamoz/internal/uuid"
)

var seedAuthors = []struct {
	nickname string
	email    string
	bio      string
	language string
	role     author.Role
}{
	{nickname: "yakamoz", email: "yakamoz@example.com", bio: "Gece denizin üzerinde ay ışığının bıraktığı gümüş yansıma.", language: "tr", role: author.RoleAuthor},
	{nickname: "geceyazari", email: "gece@example.com", bio: "Night owl writing about cities, books and sea.", language: "en", role: author.RoleAuthor},
	{nickname: "marti", email: "marti@example.com", bio: "İstanbul'un tepelerinde dolaşan bir martı gibi yazıyorum.", language: "tr", role: author.RoleAuthor},
	{nickname: "elif", email: "elif@example.com", bio: "Şehirler, kahve ve küçük keşifler üzerine yazıyorum.", language: "tr", role: author.RoleAuthor},
	{nickname: "deniz", email: "deniz@example.com", bio: "Stories from the coast and beyond.", language: "en", role: author.RoleAuthor},
	{nickname: "atlas", email: "atlas@example.com", bio: "Reisen, Kultur und gute Gespräche.", language: "de", role: author.RoleAuthor},
	{nickname: "selin", email: "selin@example.com", bio: "Les livres et les lieux qui nous changent.", language: "fr", role: author.RoleAuthor},
	{nickname: "emre", email: "emre@example.com", bio: "Merak ettiklerimi paylaşmayı seviyorum.", language: "tr", role: author.RoleAuthor},
	{nickname: "maya", email: "maya@example.com", bio: "Ideas, observations, and everyday life.", language: "en", role: author.RoleAuthor},
	{nickname: "leo", email: "leo@example.com", bio: "Historias sobre cultura y viajes.", language: "es", role: author.RoleAuthor},
	{nickname: "reviewer", email: "reviewer@example.com", bio: "Reviews community submissions.", language: "en", role: author.RoleReviewer},
	{nickname: "admin", email: "admin@example.com", bio: "Manages the Yakamoz community.", language: "en", role: author.RoleAdmin},
}

var seedTopics = []struct {
	title       string
	description string
	language    string
	publish     bool
}{
	{title: "yakamoz", description: "Ay ışığının suda bıraktığı gümüşsü yansıma. Türkçenin en güzel kelimelerinden biri olarak kabul edilir ve karşılığı diğer dillerde tek kelimeyle bulunmaz.", language: "tr", publish: true},
	{title: "çay", description: "Türk kültürünün vazgeçilmez içeceği. İnce belli bardakta, tavşan kanı kıvamında, demli olanı makbuldür.", language: "tr", publish: true},
	{title: "istanbul boğazı", description: "İki kıtayı ayıran ve birleştiren su yolu. Vapur yolculukları, martılar ve simit eşliğinde dünyanın en güzel manzaralarından biri.", language: "tr", publish: true},
	{title: "gece yarısı kodlamak", description: "Herkes uyuduğunda açılan ikinci ekran. Saat 03:00'te çözülen bug'ların gizemli dünyası.", language: "tr", publish: true},
	{title: "kapadokya", description: "Peri bacaları, sıcak hava balonları ve yeraltı şehirleriyle dünyada eşi benzeri olmayan bir coğrafya.", language: "tr", publish: true},
	{title: "göbeklitepe", description: "İnsanlık tarihinin bilinen en eski tapınağı. 12.000 yıl önce inşa edilmiş, tarih kitaplarını yeniden yazdıran keşif.", language: "tr", publish: true},
	{title: "simit", description: "Susamla kaplı, dışı çıtır içi yumuşak halka ekmek. Vapurda martılara atmak için de idealdir.", language: "tr", publish: true},
	{title: "dolmuş", description: "Ücretini öndeki yolcuya uzatıp 'şurada iner misiniz?' diye seslendiğiniz efsane toplu taşıma aracı.", language: "tr", publish: true},
	{title: "vapur yolculuğu", description: "Kadıköy-Karaköy hattında 20 dakikalık meditasyon. Çay, simit, martılar ve Boğaz'ın serin rüzgarı.", language: "tr", publish: true},
	{title: "deniz feneri", description: "Karanlıkta gemilere yol gösteren yalnız bekçi. Edebiyatın en romantik metaforlarından biri.", language: "tr", publish: true},
	{title: "retro oyunlar", description: "8-bit müzikler, şifreli seviyeleri ve cartridge üfleyerek çalıştırma ritüelleriyle altın çağ.", language: "tr", publish: true},
	{title: "sahaf", description: "Tozlu raflar arasında kaybolmak. İkinci el kitabın içinde bulunan eski notların ve biletlerin büyüsü.", language: "tr", publish: true},
	{title: "yalnızlık", description: "Kalabalık bir şehirde kendi sesini duyabilmek. Bazen bir seçim, bazen bir zorunluluk.", language: "tr", publish: true},
	{title: "uzay", description: "Milyarlarca yıldız, sonsuz olasılıklar. İnsanlığın en büyük merak konusu ve en büyük bilinmezi.", language: "tr", publish: true},
	{title: "kahve", description: "Türk kahvesi bir fincanı kırk yıl hatıra sahiptir. Falına bakılmayanı eksik kalır.", language: "tr", publish: true},
	{title: "nostalji", description: "Siyah beyaz fotoğrafların, eski şarkıların ve kayıp eşyaların çağrıştırdığı buruk duygu.", language: "tr", publish: true},
	{title: "moonlight on water", description: "The Turkish word 'yakamoz' describes this sight — the shimmering path moonlight draws across the sea at night.", language: "en", publish: true},
	{title: "night trains", description: "Compartments, clacking rails, and cities that appear out of the dark. The most romantic way to cross a continent.", language: "en", publish: true},
	{title: "taslak konu", description: "Henüz yayınlanmamış bir konu. Yayına hazır olduğunda admin panelinden publish edilebilir.", language: "tr", publish: false},
	{title: "kayıp eşya", description: "Bir türlü bulunamayan çorapların ve kalemlerin gittiği gizemli boyut. Kimse nereye gittiklerini bilmiyor.", language: "tr", publish: false},
}

var seedTranslations = []struct {
	topicIndex  int
	language    string
	title       string
	description string
}{
	{0, "en", "Moonlight on the sea", "Yakamoz is the silvery shimmer moonlight leaves on the sea at night."},
	{0, "de", "Mondschein auf dem Meer", "Yakamoz ist der silbrige Schimmer, den das Mondlicht nachts auf dem Meer hinterlässt."},
	{0, "fr", "Clair de lune sur la mer", "Yakamoz désigne le reflet argenté du clair de lune sur la mer pendant la nuit."},
	{0, "es", "Luz de luna sobre el mar", "Yakamoz es el reflejo plateado que deja la luz de la luna en el mar por la noche."},
	{1, "en", "Turkish tea", "A beloved drink in Turkish culture, traditionally served strong in a small tulip-shaped glass."},
	{1, "de", "Türkischer Tee", "Ein beliebtes Getränk der türkischen Kultur, traditionell stark in einem kleinen tulpenförmigen Glas serviert."},
	{2, "en", "The Bosphorus", "The waterway connecting two continents, best enjoyed with ferries, seagulls, and simit."},
	{2, "de", "Der Bosporus", "Die Wasserstraße zwischen zwei Kontinenten, am schönsten mit Fähren, Möwen und Simit."},
	{15, "tr", "suda ay ışığı", "Yakamoz, geceleri ay ışığının denizde oluşturduğu gümüş renkli parıltıdır."},
	{16, "tr", "gece trenleri", "Kompartımanlar, rayların ritmi ve karanlıktan çıkan şehirlerle kıtaları aşmanın romantik yolu."},
}

var seedComments = []struct {
	language string
	body     string
}{
	{"tr", "Bu konu hakkında daha önce böyle düşünmemiştim."},
	{"en", "I had never thought about this topic from that angle."},
	{"de", "Ein spannendes Thema, über das ich gern mehr erfahren würde."},
	{"fr", "Un sujet intéressant qui mérite d'être exploré davantage."},
	{"es", "Un tema interesante sobre el que me gustaría saber más."},
}

func seed(ctx context.Context, cfg config.Config, logger *slog.Logger, authorRepo author.Repository, topicRepo topic.Repository, commentRepo comment.Repository) {
	authorService := author.NewService(authorRepo)
	topicService := topic.NewService(topicRepo, ai.Stub{}, cfg.DefaultLanguage, cfg.AITimeout)
	commentService := comment.NewService(commentRepo)

	authorIDs := make([]uuid.UUID, 0, len(seedAuthors))
	for _, value := range seedAuthors {
		created, err := authorService.Create(ctx, value.nickname, value.email, value.bio, value.language, value.role)
		if err != nil {
			logger.Warn("seed author failed", "nickname", value.nickname, "error", err)
			continue
		}
		authorIDs = append(authorIDs, created.ID)
	}
	if len(authorIDs) == 0 {
		return
	}

	topicIDs := make([]uuid.UUID, len(seedTopics))
	for i, value := range seedTopics {
		created, err := topicService.Create(ctx, value.title, value.description, value.language, authorIDs[i%len(authorIDs)])
		if err != nil {
			logger.Warn("seed topic failed", "title", value.title, "error", err)
			continue
		}
		topicIDs[i] = created.ID
		if value.publish {
			if _, err := topicService.Publish(ctx, created.ID); err != nil {
				logger.Warn("seed publish failed", "title", value.title, "error", err)
			}
		}
	}

	for _, value := range seedTranslations {
		if value.topicIndex >= len(topicIDs) || topicIDs[value.topicIndex] == uuid.Nil {
			continue
		}
		if _, err := topicService.AddTranslation(ctx, topicIDs[value.topicIndex], value.language, value.title, value.description); err != nil {
			logger.Warn("seed translation failed", "topic_id", topicIDs[value.topicIndex], "language", value.language, "error", err)
		}
	}

	for i, topicID := range topicIDs {
		if topicID == uuid.Nil {
			continue
		}
		for j, value := range seedComments {
			commenter := authorIDs[(i+j+1)%len(authorIDs)]
			body := seedTopics[i].title + ": " + value.body
			if _, err := commentService.Create(ctx, topicID, commenter, value.language, body); err != nil {
				logger.Warn("seed comment failed", "topic_id", topicID, "language", value.language, "error", err)
			}
		}
	}
}
