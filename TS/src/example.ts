#!/usr/bin/env npx tsx

// Example script demonstrating the OVERTURE-Play TypeScript library

import * as fs from "fs";
import { Client, AudioFormats, Extension, BestAudioFormat } from "./index.js";
import { Handler, Progress } from "./stream/handler.js";

async function main() {
    const args = process.argv.slice(2);

    if (args.length < 1) {
        console.log("Usage: npx tsx src/example.ts <video-id-or-url>");
        process.exit(1);
    }

    const videoInput = args[0];

    // Create a new YouTube client
    const client = new Client({
        Debug: false,
    });

    // Fetch video information
    console.log(`Fetching video: ${videoInput}\n`);

    let video;
    try {
        video = await client.GetVideo(videoInput);
    } catch (err) {
        console.error(`Error fetching video: ${err}`);
        process.exit(1);
    }

    // Print video details
    console.log(`Title:       ${video.Title}`);
    console.log(`Author:      ${video.Author}`);
    console.log(`Duration:    ${video.Duration} seconds`);
    console.log(`Views:       ${video.ViewCount}\n`);

    // Print available formats
    console.log("Video Formats:");
    for (const format of video.Formats.filter((f) => f.Width > 0 && f.Height > 0)) {
        console.log(
            `  ITag: ${format.ITag.toString().padEnd(4)} | ${(format.QualityLabel || format.Quality).padEnd(10)} | ${format.Width}x${format.Height} @ ${format.FPS}fps | ${Extension(format)}`
        );
    }

    console.log("\nAudio Formats:");
    const audioFormats = AudioFormats(video);
    for (const format of audioFormats) {
        console.log(
            `  ITag: ${format.ITag.toString().padEnd(4)} | ${(format.AudioQuality || "unknown").padEnd(15)} | ${Math.floor(format.Bitrate / 1000)} kbps | ${Extension(format)}`
        );
    }

    // Find best audio format
    if (audioFormats.length === 0) {
        console.log("\nNo audio formats available");
        return;
    }

    const bestAudio = BestAudioFormat(video);
    if (!bestAudio) {
        console.log("\nNo audio formats available");
        return;
    }

    // Create safe filename
    const safeTitle = sanitizeFilename(video.Title || `video_${video.ID}`);
    const filename = `${safeTitle}.${Extension(bestAudio)}`;

    console.log("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
    console.log(`Downloading: ${filename}`);
    console.log(`Format: ITag ${bestAudio.ITag}, ${Math.floor(bestAudio.Bitrate / 1000)} kbps, ${Extension(bestAudio)}`);

    if (bestAudio.ContentLength > 0) {
        console.log(`Size: ${(bestAudio.ContentLength / 1024 / 1024).toFixed(2)} MB`);
    }

    console.log();

    // Download with progress
    const startTime = Date.now();
    const streamHandler = new Handler();

    try {
        const data = await streamHandler.DownloadWithProgress(bestAudio, (p: Progress) => {
            if (p.Total > 0) {
                const percent = (p.Downloaded / p.Total) * 100;
                const mbDownloaded = p.Downloaded / 1024 / 1024;
                const mbTotal = p.Total / 1024 / 1024;
                const speed = p.Speed / 1024 / 1024;

                process.stdout.write(
                    `\r  Progress: ${percent.toFixed(1)}% (${mbDownloaded.toFixed(2)} / ${mbTotal.toFixed(2)} MB) @ ${speed.toFixed(2)} MB/s`
                );
            } else {
                const mbDownloaded = p.Downloaded / 1024 / 1024;
                process.stdout.write(`\r  Downloaded: ${mbDownloaded.toFixed(2)} MB`);
            }
        });

        // Write to file
        fs.writeFileSync(filename, data);

        const elapsed = (Date.now() - startTime) / 1000;
        console.log();
        console.log(`\n✓ Download complete in ${elapsed.toFixed(1)} seconds`);
        console.log(`  Saved to: ${filename}`);
    } catch (err) {
        console.log();
        console.error(`\nError downloading: ${err}`);
        process.exit(1);
    }
}

function sanitizeFilename(name: string): string {
    if (!name) {
        return "download";
    }
    
    // Replace characters not allowed in filenames
    name = name
        .replace(/[/\\:*?"<>|]/g, "_")
        .trim();

    // Truncate if too long
    if (name.length > 100) {
        name = name.slice(0, 100);
    }

    return name || "download";
}

main().catch((err) => {
    console.error(err);
    process.exit(1);
});
