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
}{
	{nickname: "yakamoz", email: "yakamoz@example.com", bio: "Gece denizin üzerinde ay ışığının bıraktığı gümüş yansıma.", language: "tr"},
	{nickname: "geceyazari", email: "gece@example.com", bio: "Night owl writing about cities, books and sea.", language: "en"},
	{nickname: "marti", email: "marti@example.com", bio: "İstanbul'un tepelerinde dolaşan bir martı gibi yazıyorum.", language: "tr"},
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

func seed(ctx context.Context, cfg config.Config, logger *slog.Logger, authorRepo author.Repository, topicRepo topic.Repository, commentRepo comment.Repository) {
	authorService := author.NewService(authorRepo)
	topicService := topic.NewService(topicRepo, ai.Stub{}, cfg.DefaultLanguage, cfg.AITimeout)
	commentService := comment.NewService(commentRepo)

	authorIDs := make([]uuid.UUID, 0, len(seedAuthors))
	for _, value := range seedAuthors {
		created, err := authorService.Create(ctx, value.nickname, value.email, value.bio, value.language)
		if err != nil {
			logger.Warn("seed author failed", "nickname", value.nickname, "error", err)
			continue
		}
		authorIDs = append(authorIDs, created.ID)
	}
	if len(authorIDs) == 0 {
		return
	}

	topicIDs := make([]uuid.UUID, 0, len(seedTopics))
	for i, value := range seedTopics {
		created, err := topicService.Create(ctx, value.title, value.description, value.language, authorIDs[i%len(authorIDs)])
		if err != nil {
			logger.Warn("seed topic failed", "title", value.title, "error", err)
			continue
		}
		if value.publish {
			if _, err := topicService.Publish(ctx, created.ID); err != nil {
				logger.Warn("seed publish failed", "title", value.title, "error", err)
			}
		}
		topicIDs = append(topicIDs, created.ID)
	}

	for i, topicID := range topicIDs {
		commenter := authorIDs[(i+1)%len(authorIDs)]
		if _, err := commentService.Create(ctx, topicID, commenter, "tr", "Bu konu hakkında söylenecek çok şey var."); err != nil {
			logger.Warn("seed comment failed", "topic_id", topicID, "error", err)
		}
	}
}
