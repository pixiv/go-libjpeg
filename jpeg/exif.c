#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <setjmp.h>
#include "jpeglib.h"
#include "exif.h"

#define EXIF_JPEG_MARKER   JPEG_APP0+1
#define EXIF_IDENT_STRING  "Exif\000\000"

#define G_LITTLE_ENDIAN     1234
#define G_BIG_ENDIAN        4321

typedef unsigned int uint;
typedef unsigned short ushort;

const char leth[]  = {0x49, 0x49, 0x2a, 0x00};	// Little endian TIFF header
const char beth[]  = {0x4d, 0x4d, 0x00, 0x2a};	// Big endian TIFF header

LOCAL( unsigned short )
de_get16( void * ptr, uint endian ) {
    unsigned char *bytes = (unsigned char *)(ptr);
    if ( endian == G_BIG_ENDIAN )
    {
        return bytes[1] + (((uint)bytes[0]) << 8);
    }
    return bytes[0] + (((uint)bytes[1]) << 8);
}

LOCAL( unsigned int )
de_get32( void * ptr, uint endian ) {
    unsigned char *bytes = (unsigned char *)(ptr);
    if ( endian == G_BIG_ENDIAN )
    {
        return bytes[3] + (((uint)bytes[2]) << 8) + (((uint)bytes[1]) << 16) + (((uint)bytes[0]) << 24);
    }
    return bytes[0] + (((uint)bytes[1]) << 8) + (((uint)bytes[2]) << 16) + (((uint)bytes[3]) << 24);
}

int jpegtran_get_orientation (j_decompress_ptr cinfo)
{
    /* This function looks through the meta data in the libjpeg decompress structure to
       determine if an EXIF Orientation tag is present and if so return its value (1-8).
       If no EXIF Orientation tag is found 0 (zero) is returned. */

    uint   i;              /* index into working buffer */
    uint   orient_tag_id;  /* endianed version of orientation tag ID */
    uint   ret;            /* Return value */
    uint   offset;         /* de-endianed offset in various situations */
    uint   tags;           /* number of tags in current ifd */
    uint   tag;            /* de-endianed tag */
    uint   type;           /* de-endianed type of tag used as index into types[] */
    uint   count;          /* de-endianed count of elements in a tag */
    uint   tiff = 0;   	   /* offset to active tiff header */
    uint   endian = 0;     /* detected endian of data */

    jpeg_saved_marker_ptr exif_marker;  /* Location of the Exif APP1 marker */
    jpeg_saved_marker_ptr cmarker;      /* Location to check for Exif APP1 marker */

    /* check for Exif marker (also called the APP1 marker) */
    exif_marker = NULL;
    cmarker = cinfo->marker_list;
    while (cmarker) {
        if (cmarker->marker == EXIF_JPEG_MARKER) {
            /* The Exif APP1 marker should contain a unique
                           identification string ("Exif\0\0"). Check for it. */
            if (!memcmp (cmarker->data, EXIF_IDENT_STRING, 6)) {
                exif_marker = cmarker;
            }
        }
        cmarker = cmarker->next;
    }
    /* Did we find the Exif APP1 marker? */
    if (exif_marker == NULL)
        return 0;
    /* Do we have enough data? */
    if (exif_marker->data_length < 32)
        return 0;

    /* Check for TIFF header and catch endianess */
    i = 0;

    /* Just skip data until TIFF header - it should be within 16 bytes from marker start.
       Normal structure relative to APP1 marker -
            0x0000: APP1 marker entry = 2 bytes
            0x0002: APP1 length entry = 2 bytes
            0x0004: Exif Identifier entry = 6 bytes
            0x000A: Start of TIFF header (Byte order entry) - 4 bytes
                    - This is what we look for, to determine endianess.
            0x000E: 0th IFD offset pointer - 4 bytes

            exif_marker->data points to the first data after the APP1 marker
            and length entries, which is the exif identification string.
            The TIFF header should thus normally be found at i=6, below,
            and the pointer to IFD0 will be at 6+4 = 10.
        */

    while (i < 16) {

        /* Little endian TIFF header */
        if (memcmp (&exif_marker->data[i], leth, 4) == 0){
            endian = G_LITTLE_ENDIAN;
        }

        /* Big endian TIFF header */
        else if (memcmp (&exif_marker->data[i], beth, 4) == 0){
            endian = G_BIG_ENDIAN;
        }

        /* Keep looking through buffer */
        else {
            i++;
            continue;
        }
        /* We have found either big or little endian TIFF header */
        tiff = i;
        break;
    }

    /* So did we find a TIFF header or did we just hit end of buffer? */
    if (tiff == 0)
        return 0;

    /* Read out the offset pointer to IFD0 */
    offset  = de_get32(&exif_marker->data[i] + 4, endian);
    i       = i + offset;

    /* Check that we still are within the buffer and can read the tag count */
    if ((i + 2) > exif_marker->data_length)
        return 0;

    /* Find out how many tags we have in IFD0. As per the TIFF spec, the first
           two bytes of the IFD contain a count of the number of tags. */
    tags    = de_get16(&exif_marker->data[i], endian);
    i       = i + 2;

    /* Check that we still have enough data for all tags to check. The tags
           are listed in consecutive 12-byte blocks. The tag ID, type, size, and
           a pointer to the actual value, are packed into these 12 byte entries. */
    if ((i + tags * 12) > exif_marker->data_length)
        return 0;

    /* Check through IFD0 for tags of interest */
    while (tags--){
        tag    = de_get16(&exif_marker->data[i], endian);
        type   = de_get16(&exif_marker->data[i + 2], endian);
        count  = de_get32(&exif_marker->data[i + 4], endian);

        /* Is this the orientation tag? */
        if (tag == 0x112){

            /* Check that type and count fields are OK. The orientation field
                           will consist of a single (count=1) 2-byte integer (type=3). */
            if (type != 3 || count != 1) return 0;

            /* Return the orientation value. Within the 12-byte block, the
                           pointer to the actual data is at offset 8. */
            ret =  de_get16(&exif_marker->data[i + 8], endian);
            return ret <= 8 ? ret : 0;
        }
        /* move the pointer to the next 12-byte tag field. */
        i = i + 12;
    }

    return 0; /* No EXIF Orientation tag found */
}
