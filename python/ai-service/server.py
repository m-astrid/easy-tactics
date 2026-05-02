import os
import uuid
from datetime import datetime
from concurrent import futures

import grpc
from google.protobuf import timestamp_pb2
from dotenv import load_dotenv

import ai_service_pb2
import ai_service_pb2_grpc

load_dotenv()


class AIServiceServicer(ai_service_pb2_grpc.AIServiceServicer):
    def AnalyzeFight(self, request, context):
        print(f"AnalyzeFight: fighter={request.fighter_uuid}, fight={request.fight_uuid}")
        
        observations = [
            "Strong defensive stance",
            "Good tempo control",
            "Excellent counter-attack timing"
        ]
        techniques = [
            "Parry-riposte",
            "Counter-cut",
            "Fast recovery"
        ]
        
        return ai_service_pb2.AnalysisResult(
            uuid=str(uuid.uuid4()),
            fighter_uuid=request.fighter_uuid,
            observations=observations,
            techniques=techniques,
            summary=f"Analyzed fight vs {request.opponent_name}: {request.score_win}-{request.score_lose}",
            analyzed_at=_now()
        )

    def GenerateSummary(self, request, context):
        print(f"GenerateSummary: fighter={request.fighter_uuid}, fights={len(request.fight_uuids)}")
        
        content = f"""# Fighter Analysis Summary

## Overview
This fencer has shown consistent improvement in technical execution.

## Strengths
- Strong defensive technique
- Good tactical awareness
- Excellent footwork

## Areas for Improvement
- Could work on attack initiation
- Tempo variation could be improved

## Recommendations
Continue drilling counter-attacks and work on combining techniques.
"""
        
        return ai_service_pb2.SummaryResult(
            fighter_uuid=request.fighter_uuid,
            content=content,
            version=1,
            updated_at=_now()
        )

    def SearchVideos(self, request, context):
        print(f"SearchVideos: fighter={request.fighter_name}")
        
        videos = []
        if request.search_youtube:
            videos.append(ai_service_pb2.VideosResult.Video(
                platform="youtube",
                url=f"https://youtube.com/watch?v=demo123",
                title=f"{request.fighter_name} - Tournament Finals",
                duration_seconds=300,
                fight_uuid=request.fight_uuids[0] if request.fight_uuids else ""
            ))
        
        return ai_service_pb2.VideosResult(videos=videos)

    def ParseHemagon(self, request, context):
        print(f"ParseHemagon: slug={request.fighter_slug}")
        
        return ai_service_pb2.ParseResult(
            fighter_uuid=str(uuid.uuid4()),
            hemagon_url=f"https://hemagon.com/fencer/{request.fighter_slug}",
            tournaments_count=5,
            fights_count=23,
            parsed_at=_now()
        )


def _now():
    ts = timestamp_pb2.Timestamp()
    ts.GetCurrentTime()
    return ts


def serve(port=50052):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    ai_service_pb2_grpc.add_AIServiceServicer_to_server(
        AIServiceServicer(), server
    )
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    print(f"AI Service gRPC server started on port {port}")
    server.wait_for_termination()


if __name__ == '__main__':
    serve()