/**

perf tests

*/

package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/backends"
	"github.com/knights-analytics/hugot/options"
	"github.com/knights-analytics/hugot/pipelines"
)

func perftests(backend string, executionProvidor string, length int, model string, modelFile string) {

	var session *hugot.Session
	var err error

	switch backend {
	case "XLA":
		session, err = hugot.NewXLASession()
	case "ORT":
		if executionProvidor == "CoreML" {

			// https://onnxruntime.ai/docs/execution-providers/CoreML-ExecutionProvider.html
			session, err = hugot.NewORTSession(options.WithCoreML(map[string]string{"ModelFormat": "MLProgram", "MLComputeUnits": "ALL", "RequireStaticInputShapes": "0", "EnableOnSubgraphs": "0"}))

		} else if executionProvidor == "XNNPACK" {

			// https://onnxruntime.ai/docs/execution-providers/Xnnpack-ExecutionProvider.html
			session, err = hugot.NewORTSession(options.WithOnnxLibraryPath("onnx/onnxruntime/build/MacOS/RelWithDebInfo"),
				options.WithInterOpNumThreads(1),
				options.WithInterOpSpinning(false),
				options.WithExtraExecutionProvider("XNNPACK", map[string]string{"intra_op_num_threads": strconv.Itoa(runtime.NumCPU())}))

		} else if executionProvidor == "ACL" {

			// https://onnxruntime.ai/docs/execution-providers/community-maintained/ACL-ExecutionProvider.html

			session, err = hugot.NewORTSession(options.WithOnnxLibraryPath("onnx/onnxruntime/build/MacOS/RelWithDebInfo"),
				options.WithExtraExecutionProvider("ACL", map[string]string{}))
		} else {

			//session, err = hugot.NewORTSession(options.WithOnnxLibraryPath("onnx/onnxruntime/build/MacOS/RelWithDebInfo"))
			session, err = hugot.NewORTSession()

		}
		// RKNPU on linux ? https://onnxruntime.ai/docs/execution-providers/community-maintained/RKNPU-ExecutionProvider.html

	default:
		// tends to hang
		session, err = hugot.NewGoSession()
	}
	if err != nil {
		panic(fmt.Sprintf("Could not start hugot: %v", err))
	}

	downloadOptions := hugot.NewDownloadOptions()
	downloadOptions.OnnxFilePath = modelFile
	modelPath, err := hugot.DownloadModel(model, "./models/", downloadOptions)
	if err != nil {
		panic(fmt.Sprintf("could not download model: %v", err))
	}
	config := hugot.ZeroShotClassificationConfig{
		ModelPath: modelPath,
		Name:      "testPipeline",
		Options: []backends.PipelineOption[*pipelines.ZeroShotClassificationPipeline]{
			pipelines.WithLabels(append(cliOptions.Include, cliOptions.Exclude...)),
			pipelines.WithMultilabel(false),
		},
	}
	classificationPipeline, err := hugot.NewPipeline(session, config)
	if err != nil {
		panic(fmt.Sprintf("could not create pipeline: %v", err))
	}

	// test classification
	//

	fmt.Fprintf(os.Stderr, "Running pipeline %s %s %s %d\n", backend, model, executionProvidor, length)

	title := "Hope in Action: Being Human in the Age of Generative AI"
	description := "Overview\n\nRethinking creativity, work, and agency in the age of generative artificial intelligence.\n\nWe are living through a moment where tools like ChatGPT, Midjourney, and GitHub Copilot are no longer futuristic curiosities, they are genuine collaborators in our writing, coding, designing, and decision-making.\n\n\n\n\nThis talk steps back from the hype to ask a more human question: what happens to creativity, work, and personal agency when machines can generate ideas, images, and solutions at scale? Technology and privacy lawyer Maleeha Akhtar will explore how these systems blur lines we once took for granted between author and tool, employee and employer, automation and autonomy. We’ll consider what it means to create in an age of algorithmic assistance, how power shifts when data becomes raw material for intelligence, and how law can protect not just innovation, but dignity, fairness, and meaningful human choice.\n\n\n\n\nUltimately, this session is about ensuring that as AI grows more capable, we remain intentional about the kind of society, and the kind of human role within it, we want to build.\n\n\n\n\nHope in Action Lecture Series\n\n\n\n\nWith so many reasons for despair, where are we finding real cause for hope?\n\nThe Hope in Action Lecture Series from the University of St. Michael's College Continuing Education brings together innovators, leaders, alumni, and faculty who are choosing courage over cynicism. Through dynamic conversations held every six weeks, this series explores how hope becomes action in sustainability, social impact, spirituality, leadership, education, the arts, and beyond.\n\nHope is not wishful thinking. It is the decision to engage with our world’s most urgent challenges and work toward meaningful change , from climate and culture to how we live our values in our workplaces and communities.\n\nJoin us for bold ideas, practical inspiration, and living examples of radical hope in our time. Come to be inspired. Leave ready to act.\n\nRead more"
	//title := "The Platform Decay: A Discussion of \"Enshittification\" by Cory Doctorow"
	//description := "Overview\n\nJoin us for an online discussion on March 25th!\n\nWhy do the digital platforms we once loved eventually turn against us? In his 2025 book, Enshittification, Cory Doctorow explores the seemingly inevitable lifecycle of modern tech giants: first, they are good to their users; then they abuse their users to favor their business customers; finally, they abuse those customers to claw back all the value for themselves before eventually dying.\n\nJoin the Austin Forum for a provocative online book discussion on this critical framework for understanding the modern web. We will move beyond the cynicism to discuss the technical and policy \"antidotes\" that Doctorow proposes to keep the internet free, fair, and functional.\n\nOur discussion will focus on the role of the technologist in resisting platform decay:\n\nThe Lifecycle of a Platform: Understanding the economic and technical incentives that drive companies toward \"enshittification\" and how to identify the warning signs early.\nAdversarial Interoperability: Discussing the technical right to build tools that plug into existing platforms—even without their permission—as a way to return power to the users.\nBuilding for Longevity: How technologists can design systems that are \"anti-enshittification\" by default, focusing on decentralized protocols, data portability, and user-centric architecture.\nThe Austin Tech Response: How our local startup and development community can build the next generation of \"honest\" platforms that resist the urge to capture and exploit their user base.\n\nWhether you are a platform strategist, a software architect or developer, or a concerned digital citizen, join us to discuss how we can save the internet from its own worst impulses.\n\n\n\n\nAttendance Instructions\n\nThe discussion will be held online via Google Meet. Please register to receive the meet link!\n\nSpace is limited, so please register only if you’re confident you can attend—and kindly cancel your registration if your plans change so we can open your spot to another participant.\n\nRead more"

	limit := length
	words := strings.Split(title+" "+description, " ")
	if len(words) < limit {
		limit = len(words)
	}
	batch := []string{strings.Join(words[:limit], " ")}
	//fmt.Fprintf(os.Stderr, "%s\n", strings.Join(words[:limit], " "))

	start := time.Now()
	batchResult, err := classificationPipeline.RunPipeline(batch)
	elapsed := time.Since(start)
	if err != nil {
		panic(fmt.Sprintf("could not run pipeline: %v", err))
	}
	fmt.Fprintf(os.Stderr, "Done running pipeline ... took %s\n", elapsed)
	if len(batchResult.GetOutput()) == 1 {
		for i := range batchResult.ClassificationOutputs[0].SortedValues {
			if batchResult.ClassificationOutputs[0].SortedValues[i].Value > mlMinScore {
				fmt.Fprintf(os.Stderr, "%s ", batchResult.ClassificationOutputs[0].SortedValues[i].Key)
			}
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	session.Destroy()
}
