package priming

type SeedContent struct {
	Source   string
	Title    string
	Content  string
	Category string
	URL      string
}

func DefaultPrimingContent() []SeedContent {
	return []SeedContent{
		{
			Source:   "reddit_wiki",
			Title:    "Lucid Dreaming Wiki",
			Category: "beginner",
			URL:      "https://www.reddit.com/r/LucidDreaming/wiki/index",
			Content: `The Lucid Dreaming Wiki is a comprehensive resource covering everything from basic techniques to advanced practices.

Key Topics:
- Reality Testing: Learn to question your waking state throughout the day so it becomes habit in dreams
- Dream Recall: Keep a dream journal and write immediately upon waking, even if just fragments
- MILD (Mnemonic Induction): As you fall asleep, repeat "I will recognize I'm dreaming" while visualizing a recent dream
- WILD (Wake Induced): Transition directly from wakefulness to lucidity while maintaining consciousness
- WBTB (Wake Back to Bed): Wake after 5-6 hours, stay awake briefly, then return to sleep with intention
- Most beginners see results within 2-4 weeks of consistent practice.`,
		},
		{
			Source:   "reddit_faq",
			Title:    "Frequently Asked Questions",
			Category: "beginner",
			URL:      "https://www.reddit.com/r/LucidDreaming/wiki/faq",
			Content: `Common Questions Answered:

Q: How long does it take to have my first lucid dream?
A: Most people report their first lucid dream within 2-6 weeks of consistent practice.

Q: Is lucid dreaming safe?
A: Yes. It's a natural state of consciousness that occurs spontaneously in about 55% of people at least once in their lifetime.

Q: Can I get stuck in a lucid dream?
A: No. Your body will naturally wake or transition to non-lucid sleep.

Q: What's the best technique for beginners?
A: Start with reality testing combined with dream journaling. These foundational practices make other techniques more effective.`,
		},
		{
			Source:   "reddit_beginners_qa",
			Title:    "Beginner Q&A Part 1",
			Category: "beginner",
			URL:      "https://www.reddit.com/r/LucidDreaming/comments/3iplpa/beginners_qa/",
			Content: `Getting Started: Essential First Steps

1. Start a Dream Journal
   Keep it by your bed. Write anything you remember immediately upon waking. Even fragments strengthen recall.

2. Perform Reality Checks
   Ask "Am I dreaming?" 10-20 times daily. Check your hands, count fingers, try to push finger through palm.

3. Set Intent Before Sleep
   As you drift off, firmly intend to recognize when you're dreaming. Visualize yourself becoming lucid.

4. Wake Back to Bed (WBTB)
   Set alarm for 5-6 hours after bedtime. Stay awake 15-30 minutes, then return to sleep with strong intention.

5. Be Patient and Consistent
   Results compound. Missing one day isn't failure—just resume the next.`,
		},
		{
			Source:   "reddit_beginners_faq_extended",
			Title:    "Beginner FAQ Extended Part 2",
			Category: "beginner",
			URL:      "https://www.reddit.com/r/LucidDreaming/comments/4cpb6o/beginners_faq_extended/",
			Content: `Advanced Beginner Concepts

Dream Signs:
Pay attention to recurring elements in your dreams—these become triggers for lucidity. Common signs: impossible places, seeing deceased people, impossible physics, old homes/schools.

Stabilization Techniques:
When you become lucid, the dream may start fading. To stabilize:
- Rub your hands together
- Spin around slowly
- Touch objects in the dream
- Remind yourself "I'm dreaming" calmly

False Awakenings:
You may "wake up" within a dream. Always do a reality check upon waking! This is a prime opportunity for lucidity.

Sleep Paralysis:
Sometimes occurs during WILD attempts. It's harmless but can be frightening. Stay calm, focus on breathing, know it will pass.`,
		},
		{
			Source:   "reddit_myths",
			Title:    "Myths and Misconceptions",
			Category: "education",
			URL:      "https://www.reddit.com/r/LucidDreaming/comments/2o22rm/myths_and_misconceptions_about_lucid_dreaming/",
			Content: `Debunking Common Myths

MYTH: Lucid dreaming is unnatural or dangerous.
TRUTH: It's a well-documented, natural state studied at Stanford and other institutions.

MYTH: You need supplements or drugs to lucid dream.
TRUTH: While some supplements may help, they're not necessary. Most achieve lucidity through mental techniques alone.

MYTH: Lucid dreaming makes you tired.
TRUTH: You get the same rest. Lucid dreams occur during REM sleep, part of normal sleep architecture.

MYTH: You can practice skills in lucid dreams.
TRUTH: Research shows motor skills can actually be improved through mental rehearsal in lucid dreams.

MYTH: Everyone can lucid dream easily.
TRUTH: While most people can learn, it requires consistent practice and varies by individual.`,
		},
	}
}

func GetSeedContentAsStructs() []struct {
	Source   string
	Title    string
	Content  string
	Category string
	URL      string
} {
	content := DefaultPrimingContent()

	result := make([]struct {
		Source   string
		Title    string
		Content  string
		Category string
		URL      string
	}, len(content))

	for i, item := range content {
		result[i] = struct {
			Source   string
			Title    string
			Content  string
			Category string
			URL      string
		}{
			Source:   item.Source,
			Title:    item.Title,
			Content:  item.Content,
			Category: item.Category,
			URL:      item.URL,
		}
	}

	return result
}
