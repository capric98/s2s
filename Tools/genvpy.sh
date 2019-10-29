#!/bin/bash
echo -e "import vapoursynth as vs
source = r'${video}'
sub  = r'${subfile}' 

core = vs.get_core(threads=16)
core.max_cache_size = 8192

src = core.lsmas.LWLibavSource(source, format=\"yuv420p8\")\n" > ${script}.vpy

if [[]]; then
    echo -e "src = core.resize.Lanczos(src, matrix_s=\"709\", width=1280, height=720)\n" >> ${script}.vpy
fi

echo -e "subed = core.sub.TextFile(clip=src, file=sub)

subed.set_output()" >> ${script}.vpy